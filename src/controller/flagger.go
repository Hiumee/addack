package controller

import (
	"bytes"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/hiumee/addack/src/model"
)

type FlagSubmitter struct {
	flagQueue  chan model.Flag
	flags      []model.Flag
	tickFreq   time.Duration
	lastTick   time.Time
	controller *Controller
	mutex      *sync.Mutex
}

func NewFlagSubmitter(controller *Controller) *FlagSubmitter {
	return &FlagSubmitter{
		flagQueue:  make(chan model.Flag, 1000),
		tickFreq:   time.Duration(controller.Config.SendFlagTick) * time.Millisecond,
		lastTick:   time.Now(),
		controller: controller,
		mutex:      &sync.Mutex{},
	}

}

func (fs *FlagSubmitter) QueueFlag(flag model.Flag) {
	fs.flagQueue <- flag
}

func (fs *FlagSubmitter) Run(matchedFlags []model.Flag) {
	ticker := time.NewTicker(fs.tickFreq)
	defer ticker.Stop()

	go func() {
		for _, flag := range matchedFlags {
			fs.QueueFlag(flag)
		}
	}()

	flags := make([]model.Flag, 0)

	for {
		select {
		case flag := <-fs.flagQueue:
			fs.mutex.Lock()
			flags = append(flags, flag)
			if len(flags) >= int(fs.controller.Config.FlagMaxNum) {
				savedFlags := make([]model.Flag, len(flags))
				copy(savedFlags, flags)
				flags = make([]model.Flag, 0)

				go fs.submitFlags(savedFlags)
			}
			fs.mutex.Unlock()
		case <-ticker.C:
			fs.mutex.Lock()
			if len(flags) > 0 {
				savedFlags := make([]model.Flag, len(flags))
				copy(savedFlags, flags)
				flags = make([]model.Flag, 0)
				go fs.submitFlags(savedFlags)
			}
			fs.mutex.Unlock()
		}
	}
}

func (fs *FlagSubmitter) submitFlags(flags []model.Flag) {
	fs.controller.Logger.Println("FlagSubmitter", "Submitting flags")
	var output bytes.Buffer

	cmd := exec.Command("bash", "-c", fs.controller.Config.FlaggerCommand)
	cmd.Dir = fs.controller.Config.ExploitsPath

	// Write flags separated by ~ in env variable
	var flagsStringBuilder strings.Builder
	for _, flag := range flags {
		flagsStringBuilder.WriteString(flag.Flag)
		flagsStringBuilder.WriteString("~")
	}
	flagsString := flagsStringBuilder.String()

	flagsString = strings.TrimSuffix(flagsString, "~")

	cmd.Env = append(cmd.Env, "FLAGS="+flagsString)

	writer := io.Writer(&output)
	cmd.Stdout = writer

	err := cmd.Run()
	if err != nil {
		fs.controller.Logger.Println("FlagSubmitter error", "Could not submit flags", err)
		return
	}

	validFlags := strings.Split(string(output.Bytes()), "\n")
	fs.controller.Logger.Println("FlagSubmitter", "Received results for", len(validFlags)-1, "flags")

	for i, flag := range flags {
		if i >= len(validFlags) {
			break
		}
		err = fs.controller.DB.UpdateFlagStatus(flag.Id, validFlags[i])
		if err != nil {
			fs.controller.Logger.Println("FlagSubmitter error", "Could not update flag status", err)
			return
		}
	}
}

// func SendFlag(flag model.Flag, controller *Controller) {
// 	controller.Logger.Println("Flagger", "Sending flag", flag.Flag, flag.TargetId, flag.ExploitId)

// 	var output bytes.Buffer

// 	cmd := exec.Command("bash", "-c", controller.Config.FlaggerCommand)
// 	cmd.Dir = controller.Config.ExploitsPath
// 	cmd.Env = append(cmd.Env, "FLAG="+flag.Flag)
// 	writer := io.Writer(&output)
// 	cmd.Stdout = writer

// 	err := cmd.Run()
// 	if err != nil {
// 		controller.Logger.Println("Flagger error", "Could not send flag", err)
// 		return
// 	}

// 	valid := string(output.Bytes())
// 	valid = strings.TrimSpace(valid)
// 	controller.Logger.Println("Flagger", "Received result for", flag.Flag, flag.TargetId, flag.ExploitId, valid)

// 	err = controller.DB.UpdateFlagStatus(flag.Id, valid)
// 	if err != nil {
// 		controller.Logger.Println("ExploitRunner error", "Could not update flag status", err)
// 		return
// 	}
// }
