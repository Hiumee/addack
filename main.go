package main

import (
	"addack/src/controller"
	"addack/src/database"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pelletier/go-toml/v2"
)

//go:embed assets/**/* assets/js/* assets/css/output.css assets/favicon.ico templates/*
var staticContent embed.FS

func main() {
	config := readConfig()

	// Comment if you don't want to use embedded FS
	staticFS, err := fs.Sub(staticContent, "assets")
	if err != nil {
		panic(err)
	}

	database, err := database.NewDatabase(config.DatabasePath)
	if err != nil {
		panic(err)
	}
	defer database.DB.Close()

	ctrl := &controller.Controller{
		DB:     database,
		Config: config,
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
	// r.Use(gin.Logger()) - Comment out if you want to see gin logs

	r.MaxMultipartMemory = 50 << 20 // 50 MiB

	// Switch comments if you don't want to use embedded FS
	r.StaticFS("/assets", http.FS(staticFS))
	LoadHTMLFromEmbedFS(r, staticContent, "templates/*")
	// r.Static("/assets", "./assets")
	// r.LoadHTMLGlob("templates/*")

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

	r.Run(config.ListeningAddr)
}

func readConfig() *controller.Config {

	defaultRegex := "FLAG{.*}"
	regex, err := regexp.Compile(defaultRegex)
	if err != nil {
		panic(err)
	}

	// Default values
	config := &controller.Config{
		DatabasePath:   "database.db",
		FlaggerCommand: "python3 flagger.py",
		ExploitsPath:   "./exploits",
		TickTime:       10 * 1000,
		FlagRegex:      regex,
		TimeZone:       "Europe/Bucharest",
		TimeFormat:     "2006-01-02 15:04:05",
	}

	f, err := os.ReadFile("config.toml")
	if err != nil {
		log.Println("No config file found, using default values")
		return config
	}

	var data map[interface{}]interface{}

	err = toml.Unmarshal(f, &data)

	if err != nil {
		log.Println("Error reading config file, using default values")
		return config
	}

	if data["database_path"] != nil {
		switch data["database_path"].(type) {
		case string:
			config.DatabasePath = data["database_path"].(string)
		}
	}

	if data["flagger_command"] != nil {
		switch data["flagger_command"].(type) {
		case string:
			config.FlaggerCommand = data["flagger_command"].(string)
		}
	}

	if data["exploits_path"] != nil {
		switch data["exploits_path"].(type) {
		case string:
			config.ExploitsPath = data["exploits_path"].(string)
		}
	}

	if data["tick_time"] != nil {
		switch data["tick_time"].(type) {
		case string:
			length, err := strconv.ParseInt(data["tick_time"].(string), 10, 64)
			if err != nil {
				log.Println("Invalid tick time, using default value")
			} else {
				config.TickTime = length
			}
		case int64:
			config.TickTime = data["tick_time"].(int64)
		}
	}

	if data["flag_regex"] != nil {
		switch data["flag_regex"].(type) {
		case string:
			regexString := data["flag_regex"].(string)
			regex, err := regexp.Compile(regexString)
			if err != nil {
				panic(err)
			}
			config.FlagRegex = regex
		}
	}

	if data["timezone"] != nil {
		switch data["timezone"].(type) {
		case string:
			config.TimeZone = data["timezone"].(string)
		}
	}

	if data["time_format"] != nil {
		switch data["time_format"].(type) {
		case string:
			config.TimeFormat = data["time_format"].(string)
		}
	}

	if data["listening_addr"] != nil {
		switch data["listening_addr"].(type) {
		case string:
			config.ListeningAddr = data["listening_addr"].(string)
		}
	}

	return config
}
