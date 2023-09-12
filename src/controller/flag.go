package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (c *Controller) GetFlags(context *gin.Context) {
	flags, err := c.DB.GetFlags()
	if err != nil {
		SendError(context, err.Error())
		return
	}

	context.HTML(http.StatusOK, "flags", gin.H{"Flags": flags})

	return
}

func (c *Controller) GetFlag(context *gin.Context) {
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		SendError(context, "Invalid flag id")
		return
	}

	result, err := c.DB.GetFlagResult(id)
	if err != nil {
		SendError(context, err.Error())
		return
	}

	context.HTML(http.StatusOK, "flag-preview", gin.H{"Result": result})

	return
}
