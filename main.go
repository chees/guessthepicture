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

func bigscreen(w http.ResponseWriter, req *http.Request) {
	bigscreenTemplate.Execute(w, "ws://"+req.Host+"/echo")
}

func client(w http.ResponseWriter, req *http.Request) {
	// The "/" pattern matches everything, so we need to check that we're at the root here
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}
	clientTemplate.Execute(w, "ws://"+req.Host+"/echo")
}

func main() {
	http.HandleFunc("/echo", echo)
	http.HandleFunc("/bigscreen", bigscreen)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", client)
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


var clientTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<head>
	<meta charset="utf-8">
	<title>Guess the Picture</title>
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<style>
#main {
	display: none;
}
	</style>
</head>
<body>
<form id="setup">
	<label for="user">Jak masz na imię?</label>
	<input id="user" value="Christiaan">
	<input type="submit" value="OK">
</form>
<div id="main">
	<img src="static/img/button.gif" id="button">
</div>

<script>
var user;

var ws = new WebSocket('{{.}}');
ws.onopen = function(e) {
	console.log('OPEN');
}
ws.onclose = function(e) {
	console.log('CLOSE');
	ws = null;
}
ws.onmessage = function(e) {
	console.log('RESPONSE: ' + e.data);
}
ws.onerror = function(e) {
	console.log('ERROR: ' + e.data);
}

document.getElementById('setup').onsubmit = function(e) {
	e.preventDefault();
	user = document.getElementById('user').value;
	document.getElementById('setup').style.display = 'none';
	document.getElementById('main').style.display = 'block';
};

var audio = new Audio('static/sound/fart2.ogg');

document.getElementById('button').onclick = function(e) {
	if (!ws) {
		alert('Not connected');
		return;
	}
	console.log('SEND: ' + user);
	ws.send(user);
	audio.play();
};

</script>
</body>
</html>
`))
