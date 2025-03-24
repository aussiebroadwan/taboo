package main

import (
	"log"
	"net/http"

	"github.com/lcox74/tabo/backend/pkg/hub"
)

func main() {
	h := hub.NewHub()
	go h.Run()

	engine := NewEngine(h)
	go engine.run()

	h.OnConnect = func(client *hub.Client) {
		engine.SendCurrentGameState(client)
	}

	http.HandleFunc("/ws", h.ServeWs)
	RegisterFrontend()

	log.Println("api: listening on ':8080'")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
