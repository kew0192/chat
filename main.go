package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var users = make(map[*websocket.Conn]bool)  
var messages = []string{}
var broadcast = make(chan string)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	for _, mess := range messages {
		err := ws.WriteMessage(websocket.TextMessage, []byte(mess))
		if err != nil{
			log.Println(err)
		}
	}

	users[ws] = true
	for {
		var msg string
		err := ws.ReadJSON(&msg)
		if err != nil{
			log.Println(err)
		}
		broadcast <- msg
	}
}

func handleMessages(){
	for {
		msg := <-broadcast
		for user := range users{
			err := user.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil{
				log.Println(err)
			}
		}
	}
}
func main() {
	go handleMessages()
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/ws", handleConnections)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
