package main

import (
	"log"
	"net/http"
	"os" // ВАЖНО: добавить этот импорт

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
		return
	}
	defer ws.Close()

	// Отправляем историю сообщений
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
		// Сохраняем в историю
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
	
	// Раздаем статические файлы
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/ws", handleConnections)
	
	// ВАЖНО: используем порт из переменной окружения
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Сервер запущен на порту %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
