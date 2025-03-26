package main

import (
	"log"
	"net/http"
	"time"

	"github.com/aussiebroadwan/taboo/backend/pkg/hub"
	"github.com/aussiebroadwan/taboo/backend/pkg/web"
)

func main() {
	h := hub.NewHub()
	go h.Run()

	engine := NewEngine(h)
	go engine.run()

	h.OnConnect = func(client *hub.Client) {
		engine.SendCurrentGameState(client)
	}

	router := http.NewServeMux()

	router.HandleFunc("/ws", h.ServeWs)

	web.RegisterAPI(router)
	web.RegisterSwagger(router)
	web.RegisterFrontend(router)

	server := http.Server{
		Addr:              "0.0.0.0:8080",
		ReadHeaderTimeout: 5 * time.Second,
		Handler: web.Use(
			web.WithRecoverer,
			web.WithCORS,
			web.WithLogging,
		)(router),
	}

	log.Fatal(server.ListenAndServe())
}
