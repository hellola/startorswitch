package manager

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/hellola/startorswitch/config"
	"github.com/hellola/startorswitch/wm"
)

// Command represents a command to be executed by the manager
type Command struct {
	Mode    string
	Name    string
	Options map[string]string
}

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

// Go processes the command
func (m *Manager) Go(cmd Command) error {
	if cmd.Mode == "r" || cmd.Mode == "reset" {
		return m.StateMgr.ResetAll()
	}

	var windowType WindowType
	var switchTo bool

	switch cmd.Mode {
	case "f", "focus":
		windowType = TypeFocused
	case "a", "application":
		windowType = TypeApplication
	case "c", "clean":
		windowType = TypeClean
	case "h", "hide":
		windowType = TypeHide
		return m.HideTrackedFocused()
	case "hl", "hide-latest":
		windowType = TypeHideLatest
		return m.HideOrShowLatest()
	case "ha", "hide-all":
		windowType = TypeHideAll
		return m.HideAllTracked()
	case "s", "show-all":
		windowType = TypeShowAll
		return m.ShowAllHidden()
	default:
		return fmt.Errorf("unknown command: %s", cmd.Mode)
	}

	if cmd.Mode != "ha" && cmd.Mode != "s" && cmd.Name == "" {
		return fmt.Errorf("name is required for mode: %s", cmd.Mode)
	}

	switchTo = cmd.Options["switch_to"] == "true"

	tracked := NewTracked(cmd.Name, windowType, switchTo, m.StateMgr, m.WM)

	if windowType == TypeClean {
		return tracked.Destroy()
	}

	if err := tracked.SetupTracking(); err != nil {
		return err
	}

	if err := tracked.ShowOrHide(); err != nil {
		return err
	}

	return m.HandleOptions(tracked.State(), tracked.ID(), cmd.Options)
}

// HandleOptions processes command options
func (m *Manager) HandleOptions(state WindowState, nodeID string, options map[string]string) error {
	for key, value := range options {
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
