package controller

import (
	"addack/src/database"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	DB *database.Database
}

func (c *Controller) GetIndex(context *gin.Context) {
	context.HTML(http.StatusOK, "index", gin.H{})
	return
}

func SendError(context *gin.Context, err string) {
	context.Header("HX-Retarget", "#blackhole")
	context.Header("HX-Reswap", "innerHTML")
	context.HTML(http.StatusOK, "error", gin.H{"error": err})
}
