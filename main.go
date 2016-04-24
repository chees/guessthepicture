package main

import (
	"log"
	"net/http"
	"github.com/gorilla/websocket"
	"html/template"
	"time"
)

var controlTempl = template.Must(template.ParseFiles("templates/control.html"))
var bigscreenTempl = template.Must(template.ParseFiles("templates/bigscreen.html"))
var clientTempl = template.Must(template.ParseFiles("templates/client.html"))

var upgrader = websocket.Upgrader{}
var sendBuffer = make(chan string, 64)

func bigscreenws(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	go writePump(c)
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, []byte("OK"))
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
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		sendBuffer <- string(message[:])
	}
}

// writePump pumps messages from the clients to the bigscreen connection.
func writePump(c *websocket.Conn) {
	ticker := time.NewTicker(1 * time.Second)
	defer func() {
		ticker.Stop()
		c.Close()
	}()
	for {
		select {
		case message, ok := <-sendBuffer:
			if !ok {
				c.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func control(w http.ResponseWriter, req *http.Request) {
	controlTempl.Execute(w, "ws://"+req.Host+"/clientws")
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
	http.HandleFunc("/control", control)
	http.HandleFunc("/bigscreenws", bigscreenws)
	http.HandleFunc("/bigscreen", bigscreen)
	http.HandleFunc("/clientws", clientws)
	http.HandleFunc("/", client)
	log.Fatal(http.ListenAndServe(":8080", nil))
	log.Println("Listening on port 8080")
}
