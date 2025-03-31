package manager

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/hellola/startorswitch/config"
	"github.com/hellola/startorswitch/wm"
)

// Manager handles the main application logic
type Manager struct {
	StateMgr StateManagement
	WM       wm.WMIntegration
	Config   *config.Config
}

// NewManager creates a new Manager instance
func NewManager(cfg *config.Config, wmIntegration wm.WMIntegration) (*Manager, error) {
	stateMgr, err := NewRedisStateManagement(cfg.RedisAddr)
	if err != nil {
		return nil, err
	}

	return &Manager{
		StateMgr: stateMgr,
		WM:       wmIntegration,
		Config:   cfg,
	}, nil
}

// Go processes the command line arguments
func (m *Manager) Go(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("need a name")
	}

	if len(args) == 1 && args[0] == "r" {
		return m.StateMgr.ResetAll()
	}

	var windowType WindowType
	var name string
	var options []string
	var switchTo bool

	switch args[0] {
	case "f":
		windowType = TypeFocused
	case "a":
		windowType = TypeApplication
	case "c":
		windowType = TypeClean
	case "h":
		windowType = TypeHide
		return m.HideTrackedFocused()
	case "hl":
		windowType = TypeHideLatest
		return m.HideOrShowLatest()
	case "ha":
		windowType = TypeHideAll
		return m.HideAllTracked()
	case "s":
		windowType = TypeShowAll
		return m.ShowAllHidden()
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}

	fmt.Println("command args, windowType:", windowType)

	if len(args) > 1 {
		name = args[1]
		options = args[2:]
		switchTo = strings.Contains(strings.Join(options, " "), "switch_to")
	}

	tracked := NewTracked(name, windowType, switchTo, m.StateMgr, m.WM)
	fmt.Println("tracked!: ", tracked.Name, "options: ", options)

	if windowType == TypeClean {
		return tracked.Destroy()
	}

	if err := tracked.SetupTracking(); err != nil {
		return err
	}

	fmt.Println("show or hide..")
	if err := tracked.ShowOrHide(); err != nil {
		return err
	}

	return m.HandleOptions(tracked.State(), tracked.ID(), options)
}

// HandleOptions processes command line options
func (m *Manager) HandleOptions(state WindowState, nodeID string, options []string) error {
	for _, opt := range options {
		fmt.Println("handling options,", opt)
		parts := strings.Split(opt, "=")
		if len(parts) != 2 {
			continue
		}

		key, value := parts[0], parts[1]

		switch key {
		case "top_padding":
			cmd := "bspc config -m HDMI-1-1 top_padding "
			if state == Visible {
				cmd += "0"
			} else {
				cmd += value
			}
			return exec.Command("sh", "-c", cmd).Run()
		case "mods":
			for _, mod := range strings.Split(value, ",") {
				switch mod {
				case "sticky":
					cmd := fmt.Sprintf("bspc node %s --flag sticky", nodeID)
					if err := exec.Command("sh", "-c", cmd).Run(); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// ShowAllHidden shows all hidden windows
func (m *Manager) ShowAllHidden() error {
	hidden := m.StateMgr.AllHidden()
	for _, h := range hidden {
		tracked := NewTracked(h.Name, TypeFocused, false, m.StateMgr, m.WM)
		if err := tracked.ShowAndUpdate(); err != nil {
			return err
		}
	}
	return nil
}

// HideAllTracked hides all tracked windows
func (m *Manager) HideAllTracked() error {
	all := m.StateMgr.AllTracked()
	for name := range all {
		if name == "prev" {
			continue
		}
		tracked := NewTracked(name, TypeFocused, false, m.StateMgr, m.WM)
		if err := tracked.HideAndUpdate(); err != nil {
			return err
		}
	}
	return nil
}

// HideOrShowLatest toggles the latest window
func (m *Manager) HideOrShowLatest() error {
	latest, err := m.StateMgr.LatestShown("")
	if err != nil {
		return err
	}
	if len(latest) == 0 {
		return nil
	}
	tracked := NewTracked(latest, TypeFocused, false, m.StateMgr, m.WM)
	return tracked.ToggleAndUpdate()
}

// HideTrackedFocused hides the currently focused tracked window
func (m *Manager) HideTrackedFocused() error {
	focused := m.WM.GetFocusedID()
	all := m.StateMgr.AllTracked()
	for name, id := range all {
		if name == "prev" {
			continue
		}
		if id == focused {
			tracked := NewTracked(name, TypeFocused, false, m.StateMgr, m.WM)
			return tracked.HideAndUpdate()
		}
	}
	return nil
}
