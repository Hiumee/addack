package controller

import (
	"bytes"
	"io"
	"os/exec"
	"strings"

	"github.com/hiumee/addack/src/model"
)

func SendFlag(flag model.Flag, controller *Controller) {
	controller.Logger.Println("Flagger", "Sending flag", flag.Flag, flag.TargetId, flag.ExploitId)

	var output bytes.Buffer

	cmd := exec.Command("bash", "-c", controller.Config.FlaggerCommand)
	cmd.Dir = controller.Config.ExploitsPath
	cmd.Env = append(cmd.Env, "FLAG="+flag.Flag)
	writer := io.Writer(&output)
	cmd.Stdout = writer

	err := cmd.Run()
	if err != nil {
		controller.Logger.Println("Flagger error", "Could not send flag", err)
		return
	}

	valid := string(output.Bytes())
	valid = strings.TrimSpace(valid)
	controller.Logger.Println("Flagger", "Received result for", flag.Flag, flag.TargetId, flag.ExploitId, valid)

	err = controller.DB.UpdateFlagStatus(flag.Id, valid)
	if err != nil {
		controller.Logger.Println("ExploitRunner error", "Could not update flag status", err)
		return
	}
}
