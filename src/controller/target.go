package controller

import (
	"addack/src/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (c *Controller) GetTargets(context *gin.Context) {
	targets, err := c.DB.GetTargets()
	if err != nil {
		SendError(context, err.Error())
		return
	}

	context.HTML(http.StatusOK, "targets", gin.H{"Targets": targets})

	return
}

func (c *Controller) CreateTarget(context *gin.Context) {
	var target model.Target

	target.Name = context.PostForm("name")
	target.Ip = context.PostForm("ip")
	target.Tag = context.PostForm("tag")

	if target.Name == "" || target.Ip == "" {
		SendError(context, "Name and IP fields must be filled out")
		return
	}

	id, err := c.DB.CreateTarget(target)
	if err != nil {
		SendError(context, err.Error())
		return
	}

	context.HTML(http.StatusOK, "target-row-new", gin.H{"Name": target.Name, "Id": id, "Ip": target.Ip, "Notice": "Target created", "Tag": target.Tag})
	return
}

func (c *Controller) DeleteTarget(context *gin.Context) {
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		SendError(context, "Invalid target id")
		return
	}

	err = c.DB.DeleteTarget(id)
	if err != nil {
		SendError(context, err.Error())
		return
	}

	context.HTML(http.StatusOK, "notice", gin.H{"Notice": "Target deleted"})
	return
}

func (c *Controller) DeleteAllTargets(context *gin.Context) {
	err := c.DB.DeleteAllTargets()
	if err != nil {
		SendError(context, err.Error())
		return
	}

	c.GetTargets(context)
	return
}
