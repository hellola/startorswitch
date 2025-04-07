package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hellola/startorswitch/config"
	"github.com/hellola/startorswitch/manager"
	"github.com/hellola/startorswitch/wm"
)

// DiscardWriter is a custom writer that discards all output
type DiscardWriter struct{}

func (w *DiscardWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func main() {
	// Define flags
	mode := flag.String("mode", "", "Mode of operation (f/focus, a/application, c/clean, h/hide, hl/hide-latest, ha/hide-all, s/show-all, r/reset)")
	name := flag.String("name", "", "Name of the window/application")
	options := flag.String("options", "", "Additional options (comma-separated)")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	// Set up logging based on verbose flag
	if !*verbose {
		log.SetOutput(&DiscardWriter{})
	} else {
		log.SetOutput(os.Stderr)
	}

	// Handle reset mode
	if *mode == "r" {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		wmFactory := wm.NewFactory()
		wmIntegration, err := wmFactory.CreateWM(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating window manager: %v\n", err)
			os.Exit(1)
		}

		m, err := manager.NewManager(cfg, wmIntegration)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if err := m.StateMgr.ResetAll(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Validate required flags
	if *mode == "" {
		fmt.Fprintf(os.Stderr, "Error: mode is required\n")
		flag.Usage()
		os.Exit(1)
	}

	if *name == "" && *mode != "ha" && *mode != "s" {
		fmt.Fprintf(os.Stderr, "Error: name is required for this mode\n")
		flag.Usage()
		os.Exit(1)
	}

	// Load config and create manager
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	wmFactory := wm.NewFactory()
	wmIntegration, err := wmFactory.CreateWM(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating window manager: %v\n", err)
		os.Exit(1)
	}

	m, err := manager.NewManager(cfg, wmIntegration)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Parse options into map
	optionsMap := make(map[string]string)
	if *options != "" {
		for _, opt := range strings.Split(*options, ",") {
			parts := strings.Split(opt, "=")
			if len(parts) == 2 {
				optionsMap[parts[0]] = parts[1]
			} else {
				optionsMap[opt] = "true"
			}
		}
	}

	// Create command and execute
	cmd := manager.Command{
		Mode:    *mode,
		Name:    *name,
		Options: optionsMap,
	}

	if err := m.Go(cmd); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
