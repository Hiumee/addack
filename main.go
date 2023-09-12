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

	ctrl := &controller.Controller{
		DB:  database,
		Hub: controller.NewHub(),
		Config: &controller.Config{
			ExploitsPath: "./exploits",
			TickTime:     10 * 1000,
			FlagRegex:    "FLAG{.*}",
		},
	}
	ctrl.ExploitRunner = controller.NewExploitRunner(ctrl)

	go ctrl.ExploitRunner.Run()

	{
		exploits, err := ctrl.DB.GetExploits()
		if err != nil {
			panic(err)
		}
		for _, exploit := range exploits {
			if exploit.Enabled {
				ex := exploit
				ctrl.ExploitRunner.ExploitAdder <- &ex
			}
		}
		targets, err := ctrl.DB.GetTargets()
		if err != nil {
			panic(err)
		}
		for _, target := range targets {
			if target.Enabled {
				tg := target
				ctrl.ExploitRunner.TargetAdder <- &tg
			}
		}
	}

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

	// Exploit routes
	r.GET("/exploits", ctrl.GetExploits)
	r.POST("/exploits", ctrl.CreateExploit)
	r.DELETE("/exploits", ctrl.DeleteAllExploits)
	r.DELETE("/exploit/:id", ctrl.DeleteExploit)
	r.POST("/exploit/:id/:enable", ctrl.ToggleExploit)
	// Target routes
	r.GET("/targets", ctrl.GetTargets)
	r.POST("/targets", ctrl.CreateTarget)
	r.DELETE("/targets", ctrl.DeleteAllTargets)
	r.DELETE("/target/:id", ctrl.DeleteTarget)
	r.POST("/target/:id/:enable", ctrl.ToggleTarget)
	// Flag routes
	r.GET("/flags", ctrl.GetFlags)
	r.GET("/flag/:id", ctrl.GetFlag)
	// Settings routes
	r.POST("/settings", ctrl.SaveConfig)

	// go ctrl.Hub.Run()

	r.Run("127.0.0.1:8080") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
