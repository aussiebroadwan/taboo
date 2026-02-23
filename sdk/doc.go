// Package sdk provides a Go client for the Taboo API.
//
// # REST Client
//
// Use [Client] to interact with the REST API:
//
//	client := sdk.NewClient("http://localhost:8080",
//	    sdk.WithTimeout(10*time.Second),
//	)
//
//	// List games with pagination
//	resp, err := client.ListGames(ctx, &sdk.ListGamesOptions{
//	    Cursor: sdk.Ptr(int64(100)),
//	    Limit:  sdk.Ptr(20),
//	})
//
//	// Get a single game
//	game, err := client.GetGame(ctx, 123)
//
// # SSE Client
//
// Use [SSEClient] for real-time event streaming with a handler:
//
//	type MyHandler struct { sdk.BaseEventHandler }
//
//	func (h *MyHandler) OnGamePick(e sdk.GamePickEvent) {
//	    fmt.Printf("Pick: %d\n", e.Pick)
//	}
//
//	sse := sdk.NewSSEClient("http://localhost:8080", &MyHandler{},
//	    sdk.WithReconnectDelay(time.Second),
//	)
//	sse.Connect(ctx) // blocks, auto-reconnects
//
// Or use the channel-based handler:
//
//	handler := sdk.NewChannelHandler(100)
//	sse := sdk.NewSSEClient("http://localhost:8080", handler)
//	go sse.Connect(ctx)
//
//	for event := range handler.Events() {
//	    switch e := event.(type) {
//	    case sdk.GamePickEvent:
//	        fmt.Printf("Pick: %d\n", e.Pick)
//	    }
//	}
package sdk

// Ptr returns a pointer to the given value. Useful for optional parameters.
func Ptr[T any](v T) *T {
	return &v
}
