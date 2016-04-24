package main

import (
	"log"
	"net/http"
	"github.com/gorilla/websocket"
	"html/template"
)

var bigscreenTempl = template.Must(template.ParseFiles("templates/bigscreen.html"))
var clientTempl = template.Must(template.ParseFiles("templates/client.html"))

var upgrader = websocket.Upgrader{}
var bigConn *websocket.Conn

func bigscreenws(w http.ResponseWriter, r *http.Request) {
	bigConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer bigConn.Close()
	for {
		mt, message, err := bigConn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = bigConn.WriteMessage(mt, []byte("OK"))
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func clientws(w http.ResponseWriter, r *http.Request) {
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
		err = bigConn.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
		err = c.WriteMessage(mt, []byte("OK"))
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func bigscreen(w http.ResponseWriter, req *http.Request) {
	bigscreenTempl.Execute(w, "ws://"+req.Host+"/bigscreenws")
}

func client(w http.ResponseWriter, req *http.Request) {
	// The "/" pattern matches everything, so we need to check that we're at the root here
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}
	clientTempl.Execute(w, "ws://"+req.Host+"/clientws")
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/bigscreenws", bigscreenws)
	http.HandleFunc("/bigscreen", bigscreen)
	http.HandleFunc("/clientws", clientws)
	http.HandleFunc("/", client)
	log.Fatal(http.ListenAndServe(":8080", nil))
	log.Println("Listening on port 8080")
}
