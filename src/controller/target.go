package controller

import (
	"net/http"
	"strconv"

	"github.com/hiumee/addack/src/model"

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
	target.Enabled = context.PostForm("enabled") == "on"

	if target.Name == "" || target.Ip == "" {
		SendError(context, "Name and IP fields must be filled out")
		return
	}

	id, err := c.DB.CreateTarget(target)
	if err != nil {
		SendError(context, err.Error())
		return
	}

	if target.Enabled {
		target.Id = id

		tg := target
		c.ExploitRunner.TargetAdder <- &tg
	}

	context.HTML(http.StatusOK, "target-row-new", gin.H{"Name": target.Name, "Id": id, "Ip": target.Ip, "Notice": "Target created", "Tag": target.Tag, "Enabled": target.Enabled, "Flags": 0})
	return
}

func (c *Controller) DeleteTarget(context *gin.Context) {
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		SendError(context, "Invalid target id")
		return
	}

	target, err := c.DB.GetTarget(id)
	if err != nil {
		SendError(context, err.Error())
		return
	}
	err = c.DB.DeleteTarget(id)
	if err != nil {
		SendError(context, err.Error())
		return
	}

	if target.Enabled {
		c.ExploitRunner.TargetRemover <- &model.Target{Id: id}
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

	targets := make([]*model.Target, 0)

	for _, target := range c.ExploitRunner.targets {
		targets = append(targets, target)
	}

	for _, target := range targets {
		tg := *target
		c.ExploitRunner.TargetRemover <- &tg
	}

	c.GetTargets(context)
	return
}

func (c *Controller) ToggleTarget(context *gin.Context) {
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		SendError(context, "Invalid target id")
		return
	}
	enable := context.Param("enable") == "enable"

	target, err := c.DB.GetTarget(id)
	if err != nil {
		SendError(context, err.Error())
		return
	}

	target.Enabled = enable

	err = c.DB.SetEnabledTarget(target)
	if err != nil {
		SendError(context, err.Error())
		return
	}

	switch target.Enabled {
	case true:
		c.ExploitRunner.TargetAdder <- &target
	case false:
		c.ExploitRunner.TargetRemover <- &target
	}

	c.GetTargets(context)
	return
}
