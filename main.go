package main

import (
	"addack/src/controller"
	"addack/src/database"
	"embed"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

//go:embed assets/**/* assets/js/* assets/css/output.css assets/favicon.ico templates/*
var staticContent embed.FS

func main() {
	// Uncomment if you want to use embedded assets
	// staticFS, err := fs.Sub(staticContent, "assets")
	// if err != nil {
	// 	panic(err)
	// }

	database, err := database.NewDatabase("database.db")
	if err != nil {
		panic(err)
	}
	defer database.DB.Close()

	ctrl := &controller.Controller{
		DB: database,
		Config: &controller.Config{
			ExploitsPath: "./exploits",
			TickTime:     10 * 1000,
			FlagRegex:    "FLAG{.*}",
			TimeZone:     "Europe/Bucharest",
			TimeFormat:   "2006-01-02 15:04:05",
		},
		Logger: log.New(os.Stdout, "[ExploitRunner] ", log.LstdFlags),
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

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(
		gin.Recovery(),
	)

	r.MaxMultipartMemory = 50 << 20 // 50 MiB

	// Switch comments if you want to use embedded assets
	// r.StaticFS("/assets", http.FS(staticFS))
	// LoadHTMLFromEmbedFS(r, staticContent, "templates/*")
	r.Static("/assets", "./assets")
	r.LoadHTMLGlob("templates/*")

	r.GET("/", ctrl.GetIndex)
	r.GET("/main", ctrl.GetMain)
	r.GET("/settings", ctrl.GetSettings)

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
	r.POST("/flags", ctrl.SearchFlags)
	r.GET("/flag/:id", ctrl.GetFlag)
	// Settings routes
	r.POST("/settings", ctrl.SaveConfig)

	// go ctrl.Hub.Run()

	r.Run("127.0.0.1:8080") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
