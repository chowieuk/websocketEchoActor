package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/gorilla/websocket"
	"github.com/lmittmann/tint"
	"google.golang.org/protobuf/proto"

	"github.com/chowieuk/protoactor-go-playground/websocketEchoActor/protos"
)

var (
	addr = flag.String("addr", "localhost:8080", "http service address")

	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Not recommended for production
		},
	}

	system = actor.NewActorSystemWithConfig(
		actor.Configure(actor.WithLoggerFactory(func(system *actor.ActorSystem) *slog.Logger {
			w := os.Stderr

			// create a new logger
			return slog.New(tint.NewHandler(w, &tint.Options{
				Level:      slog.LevelInfo,
				TimeFormat: time.Kitchen,
			}))
		})))
)

type (
	GatewayActor struct {
		conn *websocket.Conn
	}
	EchoActor struct {
		conn *websocket.Conn
	}
)

func (state *GatewayActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Stopped:
		context.ActorSystem().Logger().Info("WebSocket actor stopped", slog.String("pid", context.Self().Id))
		state.cleanup()
	case *actor.Started:
		context.ActorSystem().Logger().Info("WebSocket actor started", slog.String("pid", context.Self().Id))
	case *protos.EchoRequest:
		context.ActorSystem().Logger().Info("Received echo request", slog.String("msg", fmt.Sprint(msg.Message)), slog.String("dur", fmt.Sprint(msg.Duration)))

		echoActorProps := actor.PropsFromProducer(func() actor.Actor {
			return &EchoActor{conn: state.conn}
		})
		echoActor := context.SpawnPrefix(echoActorProps, "echo")
		context.Forward(echoActor)
	}
}

func (state *GatewayActor) cleanup() {
	// Perform any necessary cleanup
	if state.conn != nil {
		state.conn.Close()
	}
}

func (state *EchoActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		context.ActorSystem().Logger().Info("Echo actor started", slog.String("pid", context.Self().Id))
	case *protos.EchoRequest:
		err := state.SimulateWork(msg.Duration)
		if err != nil {
			err = state.conn.WriteMessage(1, []byte(err.Error()))
			if err != nil {
				log.Println("write:", err)
				break
			}
			break
		}
		err = state.conn.WriteMessage(1, []byte(fmt.Sprintf("Actor %v waited %s, message: %v", context.Self().Id, msg.Duration, msg.Message)))
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func (state *EchoActor) SimulateWork(d string) error {
	duration, err := time.ParseDuration(d)
	if err != nil {
		return err
	}
	time.Sleep(duration)
	return nil
}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	gatewayActorProps := actor.PropsFromProducer(func() actor.Actor {
		return &GatewayActor{conn: c}
	})

	gatewayActor := system.Root.SpawnPrefix(gatewayActorProps, "gateway")
	defer c.Close()
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			system.Root.Stop(gatewayActor)
			break
		}

		// Deserialize the received message
		var echoReq protos.EchoRequest
		err = proto.Unmarshal(message, &echoReq)
		if err != nil {
			log.Println("unmarshal:", err)
			break
		}

		system.Root.Send(gatewayActor, &echoReq)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	fs := http.FileServer(http.Dir("dist"))
	http.Handle("/dist/", http.StripPrefix("/dist/", fs))

	http.HandleFunc("/echo", echo)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(*addr, nil))

}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script src="dist/bundle.js"></script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server, 
"Send" to send a message to the server and "Close" to close the connection. 
You can change the duration and message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p>
<label for="duration">Duration:</label>
<input id="duration" value="1s">
</p>
<p>
<label for="input">Message:</label>
<input id="input" type="text" value="Hello world!">
</p>
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output" style="max-height: 70vh;overflow-y: scroll;"></div>
</td></tr></table>
</body>
</html>
`))
