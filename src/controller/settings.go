package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (c *Controller) SaveConfig(context *gin.Context) {
	flagRegex := context.PostForm("flagRegex")
	if flagRegex == "" {
		SendError(context, "Flag regex cannot be empty")
		return
	}

	tickRate, err := strconv.ParseInt(context.PostForm("tickRate"), 10, 64)
	if err != nil {
		SendError(context, err.Error())
		return
	}

	if tickRate == 0 {
		SendError(context, "Tick rate cannot be 0")
		return
	}

	timeZone := context.PostForm("timeZone")
	if timeZone == "" {
		SendError(context, "Time zone cannot be empty")
		return
	}

	timeFormat := context.PostForm("timeFormat")
	if timeFormat == "" {
		SendError(context, "Time format cannot be empty")
		return
	}

	c.Config.FlaggerCommand = context.PostForm("flaggerCommand")
	c.Config.FlagRegex = flagRegex
	c.Config.TickTime = tickRate
	c.Config.TimeZone = timeZone
	c.Config.TimeFormat = timeFormat

	context.HTML(http.StatusOK, "settings", gin.H{"Config": c.Config})
}
