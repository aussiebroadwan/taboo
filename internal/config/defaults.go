package config

import "time"

// Default returns a Config with default values.
func Default() *Config {
	return &Config{
		Environment: "development",
		Server: ServerConfig{
			Host:            "0.0.0.0",
			Port:            8080,
			ReadTimeout:     Duration(30 * time.Second),
			WriteTimeout:    Duration(30 * time.Second),
			ShutdownTimeout: Duration(10 * time.Second),
			SSEHeartbeat:    Duration(15 * time.Second),
			RequestTimeout:  Duration(30 * time.Second),
			CORSOrigins:     []string{},
			RateLimit:       100,
			RateBurst:       20,
		},
		Game: GameConfig{
			DrawDuration: Duration(90 * time.Second),
			WaitDuration: Duration(90 * time.Second),
			PickCount:    20,
			MaxNumber:    80,
		},
		Database: DatabaseConfig{
			Driver: "sqlite",
			DSN:    "taboo.db",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
		},
		Discord: DiscordConfig{
			ClientID:     "",
			ClientSecret: "",
		},
	}
}
