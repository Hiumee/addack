package controller

import (
	"addack/src/model"
	"fmt"
	"time"
)

func SendFlag(flag model.Flag, controller *Controller) {
	// TODO: Implement this function
	// This function should send the flag to the server
	// The `valid` field of the flag should be set to "valid" if the flag is valid
	// The `valid` field of the flag should be set to "invalid: <reason>" if the flag is invalid

	// Send flag to server
	fmt.Println(flag.Flag)
	// Wait for response
	time.Sleep(4 * time.Second)
	// Update flag status
	valid := "invalid: aleady submitted"
	err := controller.DB.UpdateFlagStatus(flag.Id, valid)
	if err != nil {
		controller.Logger.Println("ExploitRunner error", "Could not update flag status", err)
		return
	}
}
