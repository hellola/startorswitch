package main

import (
	"fmt"
	"os"

	"github.com/hellola/startorswitch/config"
	"github.com/hellola/startorswitch/manager"
	"github.com/hellola/startorswitch/wm"
)

func main() {
	fmt.Println("hello!!")
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(cfg)

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

	if err := m.Go(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
