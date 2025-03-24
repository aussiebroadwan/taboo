package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/lcox74/tabo/backend/pkg/hub"
)

func main() {
	fmt.Println("Hello, World")

	h := hub.NewHub()
	go h.Run()

	engine := NewEngine(h)
	go engine.run()

	h.OnConnect = func(client *hub.Client) {
		engine.SendCurrentGameState(client)
	}

	http.HandleFunc("/ws", h.ServeWs)

	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
