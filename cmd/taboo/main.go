package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/aussiebroadwan/taboo/internal/app"
)

var (
	configPath string
	logLevel   string
	verbose    bool
)

func main() {
	// Define global flags
	flag.StringVar(&configPath, "config", "./config.yaml", "config file path")
	flag.StringVar(&configPath, "c", "./config.yaml", "config file path (shorthand)")
	flag.StringVar(&logLevel, "log-level", "", "override log level (debug, info, warn, error)")
	flag.BoolVar(&verbose, "verbose", false, "shorthand for --log-level=debug")
	flag.BoolVar(&verbose, "v", false, "shorthand for --log-level=debug (shorthand)")

	flag.Usage = printUsage
	flag.Parse()

	// Subcommand dispatch
	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	var err error
	switch args[0] {
	case "serve":
		err = app.RunServe(configPath, logLevel, verbose)
	case "migrate":
		err = app.RunMigrate(configPath, args[1:])
	case "verify":
		err = app.RunVerify(configPath)
	case "version":
		app.RunVersion()
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", args[0])
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `taboo - Keno-style lottery visualization

Usage:
  taboo [flags] <command>

Commands:
  serve     Start the HTTP server
  migrate   Manage database migrations
  verify    Verify configuration and database
  version   Print version information
  help      Show this help message

Flags:
  -c, --config string      Config file path (default "./config.yaml")
  --log-level string       Override log level (debug, info, warn, error)
  -v, --verbose            Shorthand for --log-level=debug

Examples:
  taboo serve                         Start with default config
  taboo serve -c config.yaml          Start with custom config
  taboo serve --log-level debug       Start with debug logging
  taboo migrate up                    Apply all pending migrations
  taboo migrate status                Show migration status
  taboo verify                        Verify configuration and database
  taboo version                       Print version info
`)
}
