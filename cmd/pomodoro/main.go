package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rafaelkperes/pomodoro/pkg/slack"
)

func main() {
	mux := http.DefaultServeMux
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		cmd, err := slack.ParseCommand(r.Body)
		if err != nil {
			log.Printf("error reading request body: %v", err)
			return
		}
		fmt.Println(cmd.Command)
		fmt.Println(cmd.Text)
		fmt.Println(cmd.UserName)
	})
	err := http.ListenAndServe("localhost:8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
