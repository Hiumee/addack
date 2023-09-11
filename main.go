package main

import (
	"addack/src/controller"
	"addack/src/database"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	database, err := database.NewDatabase("database.db")
	if err != nil {
		panic(err)
	}
	defer database.DB.Close()

	ctrl := &controller.Controller{DB: database, Hub: controller.NewHub(), Config: &controller.Config{ExploitsPath: "./exploits"}}

	r := gin.Default()

	r.MaxMultipartMemory = 50 << 20 // 50 MiB
	r.Static("/assets", "./assets")
	r.LoadHTMLGlob("templates/*")

	r.GET("/", ctrl.GetIndex)

	// Websocket
	// r.GET("/ws", func(c *gin.Context) {
	// 	go func() {
	// 		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	// 		if err != nil {
	// 			return
	// 		}
	// 		defer conn.Close()
	// 		conn.WriteMessage(websocket.TextMessage, []byte("Howdy!"))
	// 		ctrl.Hub.Register <- conn

	// 		ticker := time.NewTicker(time.Second * 10)
	// 		defer func() {
	// 			ticker.Stop()
	// 			ctrl.Hub.Unregister <- conn
	// 		}()
	// 		for {
	// 			select {
	// 			case <-ticker.C:
	// 				err := conn.WriteMessage(websocket.PingMessage, nil)
	// 				if err != nil {
	// 					return
	// 				}
	// 			}
	// 		}
	// 	}()
	// })

	// Challenge routes
	r.GET("/challenges", ctrl.GetChallenges)
	r.POST("/challenges", ctrl.CreateChallenge)
	r.DELETE("/challenges", ctrl.DeleteAllChallenges)
	r.DELETE("/challenge/:id", ctrl.DeleteChallenge)
	// Target routes
	r.GET("/targets", ctrl.GetTargets)
	r.POST("/targets", ctrl.CreateTarget)
	r.DELETE("/targets", ctrl.DeleteAllTargets)
	r.DELETE("/target/:id", ctrl.DeleteTarget)

	// go ctrl.Hub.Run()

	r.Run("127.0.0.1:8080") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
