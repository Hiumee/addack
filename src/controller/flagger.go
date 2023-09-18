package controller

import (
	"addack/src/model"
	"fmt"
)

func SendFlag(flag model.Flag) {
	fmt.Println(flag.Flag)
}
