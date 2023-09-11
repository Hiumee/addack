package controller

import (
	"addack/src/model"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

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

	form, err := context.MultipartForm()
	if err != nil {
		fmt.Println(err)
		SendError(context, "An error occured parsing the form")
		return
	}

	challenge.Name = context.PostForm("name")
	challenge.Command = context.PostForm("command")
	cleanPath := path.Clean("/" + strings.Trim(context.PostForm("path"), "/"))
	challenge.Path = cleanPath
	challenge.Tag = context.PostForm("tag")

	if challenge.Name == "" || challenge.Command == "" || challenge.Path == "" {
		SendError(context, "All challenge fields must be filled out")
		return
	}

	id, err := c.DB.CreateChallenge(challenge)
	if err != nil {
		SendError(context, err.Error())
		return
	}

	if _, err := os.Stat(c.Config.ExploitsPath); os.IsNotExist(err) {
		os.Mkdir(c.Config.ExploitsPath, 0755)
	}

	challengePath := path.Join(c.Config.ExploitsPath, challenge.Path)

	if _, err := os.Stat(challengePath); os.IsNotExist(err) {
		os.Mkdir(challengePath, 0755)
	}

	for _, file := range form.File["files"] {
		filename := path.Base(file.Filename)
		err := context.SaveUploadedFile(file, path.Join(challengePath, filename))
		if err != nil {
			fmt.Println(err)
			SendError(context, "Could not save file")
			return
		}
	}

	context.HTML(http.StatusOK, "challenge-row-new", gin.H{"Name": challenge.Name, "Id": id, "Notice": "Challenge created", "Tag": challenge.Tag})
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
