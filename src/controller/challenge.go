package controller

import (
	"addack/src/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (c *Controller) GetChallenges(context *gin.Context) {
	challenges, err := c.DB.GetChallenges()
	if err != nil {
		context.HTML(http.StatusInternalServerError, "error", gin.H{})
		return
	}

	context.HTML(http.StatusOK, "challenges", gin.H{"Challenges": challenges})

	return
}

func (c *Controller) CreateChallenge(context *gin.Context) {
	var challenge model.Challenge

	challenge.Name = context.PostForm("name")
	challenge.Command = context.PostForm("command")
	challenge.Path = context.PostForm("path")

	if challenge.Name == "" || challenge.Command == "" || challenge.Path == "" {
		SendError(context, "All challenge fields must be filled out")
		return
	}

	id, err := c.DB.CreateChallenge(challenge)
	if err != nil {
		SendError(context, "Could not create challenge")
		return
	}

	context.HTML(http.StatusOK, "challenge-row-new", gin.H{"Name": challenge.Name, "Id": id, "Notice": "Challenge created"})
	return
}

func (c *Controller) DeleteChallenge(context *gin.Context) {
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)

	if err != nil {
		SendError(context, "Invalid ID")
		return
	}

	err = c.DB.DeleteChallenge(id)
	if err != nil {
		SendError(context, "Could not delete challenge")
		return
	}

	context.HTML(http.StatusOK, "notice", gin.H{"Notice": "Challenge deleted"})
	return
}

func (c *Controller) DeleteAllChallenges(context *gin.Context) {
	err := c.DB.DeleteAllChallenges()
	if err != nil {
		SendError(context, "Could not delete all challenges")
		return
	}

	c.GetChallenges(context)

	return
}
