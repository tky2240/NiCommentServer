package main

import (
	"log"
	"net/http"
	"nicomment/setting"
	"strconv"

	"github.com/gorilla/websocket"
)

var listen = setting.LoadSettings("./setting.yml")

// 接続されるクライアント
var clients = make(map[string]*websocket.Conn, 0)

// メッセージブロードキャストチャネル
var broadcast = make(chan Comment)

// アップグレーダ
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {

		origin := r.Header.Get("Origin")
		print(origin)
		isAllowOrigin := false
		for _, allowOrigin := range listen.AllowOrigins {
			if allowOrigin == origin {
				isAllowOrigin = true
				break
			}
		}
		return isAllowOrigin
	},
}

// メッセージ用構造体
type Comment struct {
	UserID         string  `json:"UserID"`
	Username       string  `json:"UserName"`
	Message        string  `json:"Message"`
	IsFixedComment bool    `json:"IsFixedComment"`
	R              byte    `json:"R"`
	G              byte    `json:"G"`
	B              byte    `json:"B"`
	A              float64 `json:"A"`
}

type NiCommentClient struct {
	WebSocket *websocket.Conn
	ID        string
}

func main() {
	mux := http.NewServeMux()

	// WebSocket
	mux.HandleFunc("/ws", handleConnections)
	mux.HandleFunc("/api/close", closeWebSocket)
	go handleMessages()

	err := http.ListenAndServe(listen.IP+":"+strconv.Itoa(listen.Port), mux)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func closeWebSocket(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("ID")
	if id == "" {
		w.WriteHeader(400)
		println("Invalid Accsess")
		return
	}
	if ws, ok := clients[id]; ok {
		ws.Close()
		delete(clients, id)
		w.WriteHeader(200)
		println("Success close socket")
	} else {
		w.WriteHeader(400)
		println("ID Not exsists")
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// 送られてきたGETリクエストをWebSocketにアップグレード
	id := r.URL.Query().Get("ID")
	if id == "" {
		w.WriteHeader(400)
		println("deny connection")
		return
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	println("Connected client")
	defer ws.Close()

	println(id)
	// クライアントを登録
	if oldWebSocket, ok := clients[id]; ok {
		oldWebSocket.Close()
		delete(clients, id)
	}
	clients[id] = ws
	for {
		var comment Comment
		// 新しいメッセージをJSONとして読み込み、Message構造体にマッピング
		err := ws.ReadJSON(&comment)
		if err != nil {
			log.Printf("error: %v\n", err)
			delete(clients, id)
			break
		}
		// 受け取ったメッセージをbroadcastチャネルに送る
		broadcast <- comment
	}
}

func handleMessages() {
	for {
		// broadcastチャネルからメッセージを受け取る
		comment := <-broadcast
		// 接続中の全クライアントにメッセージを送る
		for client := range clients {
			webSocket := clients[client]
			err := webSocket.WriteJSON(comment)
			if err != nil {
				log.Printf("error: %v\n", err)
				webSocket.Close()
				delete(clients, client)
			}
		}
	}
}
