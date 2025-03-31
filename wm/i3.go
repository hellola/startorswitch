package wm

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

// I3Integration implements WMIntegration for i3
type I3Integration struct{}

// NewI3Integration creates a new i3 integration
func NewI3Integration() *I3Integration {
	log.Printf("Creating new i3 integration")
	return &I3Integration{}
}

func (w *I3Integration) Show(nodeID string) error {
	log.Printf("Showing i3 window with ID: %s", nodeID)
	log.Printf("Executing i3 command: i3-msg scratchpad show [id=%s]", nodeID)
	return exec.Command("i3-msg", "scratchpad", "show", fmt.Sprintf("[id=%s]", nodeID)).Run()
}

func (w *I3Integration) Hide(nodeID string) error {
	log.Printf("Hiding i3 window with ID: %s", nodeID)
	cmd := fmt.Sprintf("windowunmap %s", nodeID)
	log.Printf("Executing i3 command: %s", cmd)
	return exec.Command("i3-msg", "move", "scratchpad", fmt.Sprintf("[id=%s]", nodeID)).Run()
	// execCmd.Env = os.Environ()
	// execCmd.Env = append(execCmd.Env, "DISPLAY=:0")
	// output, err := execCmd.Run()
}

func (w *I3Integration) StillAlive(nodeID string) bool {
	log.Printf("Checking if i3 window %s is still alive", nodeID)
	cmd := exec.Command("i3-msg", "-t", "get_tree")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error getting i3 tree: %v", err)
		return false
	}

	var tree map[string]interface{}
	if err := json.Unmarshal(output, &tree); err != nil {
		log.Printf("Error unmarshaling i3 tree: %v", err)
		return false
	}

	isAlive := w.findNodeInTree(tree, nodeID)
	log.Printf("Window %s alive status: %v", nodeID, isAlive)
	return isAlive
}

func (w *I3Integration) findNodeInTree(node map[string]interface{}, nodeID string) bool {
	if id, ok := node["id"].(int); ok {
		currentID := fmt.Sprintf("%d", int(id))
		if currentID == nodeID {
			log.Printf("Found matching node ID in tree: %s", currentID)
			return true
		}
	}

	if nodes, ok := node["nodes"].([]interface{}); ok {
		for _, n := range nodes {
			if nodeMap, ok := n.(map[string]interface{}); ok {
				if w.findNodeInTree(nodeMap, nodeID) {
					return true
				}
			}
		}
	}

	return false
}

func (w *I3Integration) Focus(nodeID string) error {
	log.Printf("Focusing i3 window with ID: %s", nodeID)
	cmd := fmt.Sprintf("[con_id=%s] focus", nodeID)
	log.Printf("Executing i3 command: %s", cmd)
	return exec.Command("i3-msg", cmd).Run()
}

func (w *I3Integration) IsFocused(nodeID string) bool {
	log.Printf("Checking if i3 window %s is focused", nodeID)
	focused := w.GetFocusedID()
	isFocused := focused == nodeID
	log.Printf("Window %s focused status: %v (focused ID: %s)", nodeID, isFocused, focused)
	return isFocused
}

func (w *I3Integration) GetFocusedID() string {
	log.Printf("Getting focused i3 window ID")
	cmd := exec.Command("i3-msg", "-t", "get_tree")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error getting i3 tree: %v", err)
		return ""
	}

	var tree map[string]interface{}
	if err := json.Unmarshal(output, &tree); err != nil {
		log.Printf("Error unmarshaling i3 tree: %v", err)
		return ""
	}

	focusedID := w.findFocusedNode(tree)
	log.Printf("Found focused window ID: %s", focusedID)
	return focusedID
}

func (w *I3Integration) findFocusedNode(node map[string]interface{}) string {
	if focused, ok := node["focused"].(bool); ok && focused {
		if id, ok := node["id"].(float64); ok {
			return fmt.Sprintf("%d", int(id))
		}
	}

	if nodes, ok := node["nodes"].([]interface{}); ok {
		for _, n := range nodes {
			if nodeMap, ok := n.(map[string]interface{}); ok {
				if id := w.findFocusedNode(nodeMap); id != "" {
					return id
				}
			}
		}
	}

	return ""
}

func (w *I3Integration) FindOrStartApplication(name string) (string, error) {
	log.Printf("Finding or starting application: %s", name)

	// First try to find existing window
	cmd := exec.Command("xdotool", "search", "--name", name)
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		ids := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(ids) > 0 {
			log.Printf("Found existing window for %s with ID: %s", name, ids[0])
			return ids[0], nil
		}
	}

	// Start the application
	log.Printf("Starting application: %s", name)
	cmd = exec.Command(name)
	if err := cmd.Start(); err != nil {
		log.Printf("Failed to start application %s: %v", name, err)
		return "", fmt.Errorf("failed to start application: %v", err)
	}

	// Wait for window to appear
	log.Printf("Waiting for window to appear...")
	for i := 0; i < 5; i++ {
		cmd := exec.Command("xdotool", "search", "--name", name)
		output, err := cmd.Output()
		if err == nil && len(output) > 0 {
			ids := strings.Split(strings.TrimSpace(string(output)), "\n")
			if len(ids) > 0 {
				log.Printf("Found window after starting %s with ID: %s", name, ids[0])
				return ids[0], nil
			}
		}
		log.Printf("Attempt %d/5: Window not found yet, waiting...", i+1)
		time.Sleep(time.Second)
	}

	log.Printf("Failed to find window for %s after starting", name)
	return "", fmt.Errorf("failed to find window after starting application")
}
