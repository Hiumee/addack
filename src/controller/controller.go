package controller

import (
	"log"
	"net/http"
	"regexp"

	"github.com/hiumee/addack/src/database"

	"github.com/gin-gonic/gin"
)

type Config struct {
	DatabasePath   string
	ExploitsPath   string
	TickTime       int64
	FlagRegex      *regexp.Regexp
	TimeZone       string
	TimeFormat     string
	FlaggerCommand string
	ListeningAddr  string
	SendFlagTick   int64
	FlagMaxNum     int64
}

type Controller struct {
	DB            *database.Database
	Config        *Config
	ExploitRunner *ExploitRunner
	Logger        *log.Logger
	FlagSubmitter *FlagSubmitter
}

func (c *Controller) GetIndex(context *gin.Context) {
	context.HTML(http.StatusOK, "index", gin.H{"Config": c.Config})
}

func (c *Controller) GetMain(context *gin.Context) {
	context.HTML(http.StatusOK, "main", gin.H{"Config": c.Config})
}

func (c *Controller) GetSettings(context *gin.Context) {
	context.HTML(http.StatusOK, "settings", gin.H{"Config": c.Config})
}

func SendError(context *gin.Context, err string) {
	context.Header("HX-Retarget", "#blackhole")
	context.Header("HX-Reswap", "innerHTML")
	context.HTML(http.StatusOK, "error", gin.H{"error": err})
}
