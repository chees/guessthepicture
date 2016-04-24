package main

import (
	"log"
	"net/http"
	"github.com/gorilla/websocket"
	"html/template"
)

var upgrader = websocket.Upgrader{}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func bigscreen(w http.ResponseWriter, r *http.Request) {
	bigscreenTemplate.Execute(w, "ws://"+r.Host+"/echo")
}

func main() {
	http.HandleFunc("/echo", echo)
	http.HandleFunc("/bigscreen", bigscreen)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	log.Fatal(http.ListenAndServe(":8080", nil))
	log.Println("Listening on port 8080")
}

var bigscreenTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<head>
	<meta charset="utf-8">
	<title>Guess the Picture</title>
</head>
<body>
<label for="myName">Jak masz na imię?</label>
<input id="myName" value="Christiaan">
<button id="ok">OK</button>
<div id="output"></div>

<script>
var output = document.getElementById('output');
var myName = document.getElementById('myName');

function print(message) {
	var d = document.createElement('div');
	d.innerHTML = message;
	output.appendChild(d);
}

var ws = new WebSocket('{{.}}');
ws.onopen = function(evt) {
	print('OPEN');
}
ws.onclose = function(evt) {
	print('CLOSE');
	ws = null;
}
ws.onmessage = function(evt) {
	print('RESPONSE: ' + evt.data);
}
ws.onerror = function(evt) {
	print('ERROR: ' + evt.data);
}

document.getElementById('ok').onclick = function(evt) {
	if (!ws) {
		return;
	}
	print('SEND: ' + myName.value);
	ws.send(myName.value);
};

</script>
</body>
</html>
`))

