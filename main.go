package main

import (
	"addack/src/controller"
	"addack/src/database"

	"github.com/gin-gonic/gin"
)

func main() {
	database, err := database.NewDatabase("database.db")
	if err != nil {
		panic(err)
	}
	defer database.DB.Close()

	ctrl := &controller.Controller{DB: database}

	r := gin.Default()
	r.Static("/assets", "./assets")
	r.LoadHTMLGlob("templates/*")

	r.GET("/", ctrl.GetIndex)

	// Challenge routes
	r.GET("/challenges", ctrl.GetChallenges)
	r.POST("/challenges", ctrl.CreateChallenge)
	r.DELETE("/challenges", ctrl.DeleteAllChallenges)
	r.DELETE("/challenge/:id", ctrl.DeleteChallenge)

	r.Run("127.0.0.1:8080") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
