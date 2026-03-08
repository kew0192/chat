package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var users = make(map[*websocket.Conn]bool)
var messages = []string{}
var broadcast = make(chan string)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		
		// Для preflight запросов
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

func corsWebSocket(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer ws.Close()

	for _, mess := range messages {
		err := ws.WriteMessage(websocket.TextMessage, []byte(mess))
		if err != nil {
			log.Println(err)
		}
	}

	users[ws] = true
	
	for {
		var msg string
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Println(err)
			delete(users, ws)
			break
		}
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		messages = append(messages, msg)
		
		for user := range users {
			err := user.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				log.Println(err)
				user.Close()
				delete(users, user)
			}
		}
	}
}

func main() {
	go handleMessages()
	
	http.Handle("/", corsMiddleware(http.FileServer(http.Dir("."))))
	
	http.HandleFunc("/ws", corsWebSocket(handleConnections))
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Сервер запущен на порту %s", port)
	log.Printf("Откройте http://localhost:%s в браузере", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
