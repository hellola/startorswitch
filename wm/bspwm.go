package wm

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// BSPWMIntegration implements WMIntegration for bspwm
type BSPWMIntegration struct{}

// NewBSPWMIntegration creates a new BSPWM integration
func NewBSPWMIntegration() *BSPWMIntegration {
	return &BSPWMIntegration{}
}

func (w *BSPWMIntegration) Show(nodeID string) error {
	cmd := fmt.Sprintf("bspc node %s --flag hidden=off --flag sticky; bspc node -f %s", nodeID, nodeID)
	return exec.Command("sh", "-c", cmd).Run()
}

func (w *BSPWMIntegration) Hide(nodeID string) error {
	cmd := fmt.Sprintf("bspc node %s --flag hidden=on --flag sticky", nodeID)
	return exec.Command("sh", "-c", cmd).Run()
}

func (w *BSPWMIntegration) StillAlive(nodeID string) bool {
	cmd := exec.Command("bspc", "query", "-N")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	nodes := strings.Split(strings.TrimSpace(string(output)), "\n")
	nodeIDInt, err := strconv.ParseInt(nodeID, 16, 64)
	if err != nil {
		return false
	}

	for _, node := range nodes {
		nodeInt, err := strconv.ParseInt(strings.TrimSpace(node), 16, 64)
		if err != nil {
			continue
		}
		if nodeInt == nodeIDInt {
			return true
		}
	}
	return false
}

func (w *BSPWMIntegration) Focus(nodeID string) error {
	cmd := fmt.Sprintf("bspc node -f %s", nodeID)
	return exec.Command("sh", "-c", cmd).Run()
}

func (w *BSPWMIntegration) IsFocused(nodeID string) bool {
	focused := w.GetFocusedID()
	return focused == nodeID
}

func (w *BSPWMIntegration) GetFocusedID() string {
	cmd := exec.Command("bspc", "query", "-N", "-n")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func (w *BSPWMIntegration) FindOrStartApplication(name string) (string, error) {
	// First try to find existing window
	cmd := exec.Command("xdotool", "search", "--name", name)
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		ids := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(ids) > 0 {
			return ids[0], nil
		}
	}

	// Start the application
	cmd = exec.Command(name)
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start application: %v", err)
	}

	// Wait for window to appear
	for i := 0; i < 5; i++ {
		cmd := exec.Command("xdotool", "search", "--name", name)
		output, err := cmd.Output()
		if err == nil && len(output) > 0 {
			ids := strings.Split(strings.TrimSpace(string(output)), "\n")
			if len(ids) > 0 {
				return ids[0], nil
			}
		}
		time.Sleep(time.Second)
	}

	return "", fmt.Errorf("failed to find window after starting application")
}
