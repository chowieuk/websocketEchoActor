import { EchoRequest } from "./protos/echo_pb.js";
window.addEventListener("load", function (evt) {
    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;

    var print = function (message) {
        var d = document.createElement("div");
        d.textContent = message;
        output.appendChild(d);
        output.scroll(0, output.scrollHeight);
    };

    document.getElementById("open").onclick = function (evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("ws://localhost:8080/echo");
        ws.onopen = function (evt) {
            print("OPEN");
        };
        ws.onclose = function (evt) {
            print("CLOSE");
            ws = null;
        };
        ws.onmessage = function (evt) {
            print("RESPONSE: " + evt.data);
        };
        ws.onerror = function (evt) {
            print("ERROR: " + evt.data);
        };
        return false;
    };

    document.getElementById("send").onclick = function (evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value + " " + duration.value);
        const data = new EchoRequest();
        console.log(EchoRequest);
        data.setMessage(input.value);
        data.setDuration(duration.value);
        data.setMt(1); // 1 is text

        ws.send(data.serializeBinary());
        return false;
    };

    document.getElementById("close").onclick = function (evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };
});
