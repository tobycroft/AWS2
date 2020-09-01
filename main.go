package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"main.go/config"
	"main.go/function/http"
	"main.go/function/ws"
	http2 "net/http"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http2.Request) bool {
		return true
	},
}

func main() {
	r := gin.Default()
	// websocket echo
	r.Any("/", func(c *gin.Context) {
		r := c.Request
		w := c.Writer
		if !websocket.IsWebSocketUpgrade(r) {
			http.Handler(c)
			return
		} else {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				fmt.Printf("err = %s\n", err)
				return
			}
			ws_handler(conn)
		}
	})
	r.Any("/test", func(c *gin.Context) {
		c.File("./html/index.html")
	})
	r.Any("/favicon.ico", func(c *gin.Context) {
		c.Abort()
	})
	r.Run(config.SERVER_LISTEN_ADDR + ":" + config.SERVER_LISTEN_PORT)
}

func ws_handler(conn *websocket.Conn) {
	defer ws.On_close(conn)
	//连入时发送欢迎消息
	go ws.On_connect(conn)
	for {
		mt, d, err := conn.ReadMessage()
		conn.RemoteAddr()
		if mt == -1 {
			break
		}
		if err != nil {
			fmt.Println(mt)
			fmt.Printf("read fail = %v\n", err)
			break
		}
		ws.Handler(string(d), conn)
	}
}
