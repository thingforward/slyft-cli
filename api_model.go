package main

import (
	"fmt"
	"net/http"
)

type SlyftApiModelInterface interface {
	EndPoint() string
	getName() string
}

func DeleteApiModel(inst SlyftApiModelInterface) {
	if inst == nil {
		return
	}
	confirm := askForConfirmation("Are you sure to delete project '" + inst.getName() + "'?")
	if confirm {
		resp, err := Do(inst.EndPoint(), "DELETE", nil)
		defer resp.Body.Close()
		if err != nil || resp.StatusCode != http.StatusNoContent {
			fmt.Printf("Something went wrong. Please try again. (ResponseCode: %d)\n", resp.StatusCode)
			Log.Debug(err)
		} else {
			fmt.Println("Was successfully deleted")
		}
	} else {
		fmt.Println("Good decision!")
	}
}
