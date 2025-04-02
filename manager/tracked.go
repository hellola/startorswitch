package manager

import (
	"errors"
	"log"

	"github.com/hellola/startorswitch/wm"
)

// Tracked represents a tracked window
type Tracked struct {
	Name     string
	Type     WindowType
	SwitchTo bool
	StateMgr StateManagement
	WM       wm.WMIntegration
}

// NewTracked creates a new Tracked instance
func NewTracked(name string, windowType WindowType, switchTo bool, stateMgr StateManagement, wm wm.WMIntegration) *Tracked {
	log.Printf("Creating new tracked window: name=%s, type=%v, switchTo=%v", name, windowType, switchTo)
	return &Tracked{
		Name:     name,
		Type:     windowType,
		SwitchTo: switchTo,
		StateMgr: stateMgr,
		WM:       wm,
	}
}

// ID returns the window ID
func (t *Tracked) ID() string {
	id := t.StateMgr.GetID(t.Name)
	log.Printf("Getting ID for window %s: %s", t.Name, id)
	return id
}

// State returns the current window state
func (t *Tracked) State() WindowState {
	state := t.StateMgr.GetState(t.ID())
	log.Printf("Getting state for window %s: %v", t.Name, state)
	if state == Errored {
		return Visible
	}
	return state
}

// SetupTracking initializes tracking for the window
func (t *Tracked) SetupTracking() error {
	log.Printf("Setting up tracking for window %s", t.Name)
	if t.Type == TypeApplication && t.IsTracked() && !t.WM.StillAlive(t.ID()) {
		log.Printf("Application window %s is no longer alive, destroying", t.Name)
		t.Destroy()
	}
	if t.IsTracked() {
		log.Printf("Window %s is already tracked", t.Name)
		return nil
	}

	var focusedID string
	if t.Type == TypeApplication {
		log.Printf("Finding or starting application %s", t.Name)
		var err error
		focusedID, err = t.WM.FindOrStartApplication(t.Name)
		if err != nil {
			log.Printf("Error finding or starting application %s: %v", t.Name, err)
			return err
		}
		log.Printf("Found/started application %s with ID %s", t.Name, focusedID)
	} else {
		focusedID = t.WM.GetFocusedID()
	}

	log.Printf("Saving current state for window %s", t.Name)
	return t.StateMgr.SaveCurrent(t.Name, t.Type, focusedID)
}

// Destroy removes the window from tracking
func (t *Tracked) Destroy() error {
	log.Printf("Destroying tracked window %s", t.Name)
	return t.StateMgr.DestroyID(t.Name)
}

// Hide hides the window
func (t *Tracked) Hide() error {
	log.Printf("Hiding window %s", t.Name)
	return t.WM.Hide(t.ID())
}

// IsTracked checks if the window is being tracked
func (t *Tracked) IsTracked() bool {
	isTracked := t.StateMgr.IsTracked(t.Name)
	log.Printf("Checking if window %s is tracked: %v", t.Name, isTracked)
	return isTracked
}

// SetState updates the window state
func (t *Tracked) SetState(state WindowState) error {
	log.Printf("Setting state for window %s to %v", t.Name, state)
	return t.StateMgr.SetState(t.Name, state)
}

// ShowAndUpdate shows the window and updates the state management
func (t *Tracked) ShowAndUpdate() error {
	log.Printf("Showing and updating window %s", t.Name)
	if err := t.WM.Show(t.ID()); err != nil {
		log.Printf("Error showing window %s: %v", t.Name, err)
		return err
	}
	if _, err := t.StateMgr.LatestShown(t.Name); err != nil {
		log.Printf("Error updating latest shown for window %s: %v", t.Name, err)
		return err
	}
	return t.SetState(Visible)
}

// HideAndUpdate hides the window and updates the state management
func (t *Tracked) HideAndUpdate() error {
	log.Printf("Hiding and updating window %s", t.Name)
	if t.StateMgr.LatestCount() > 1 {
		if err := t.StateMgr.RemoveFromLatest(t.Name); err != nil {
			log.Printf("Error removing from latest for window %s: %v", t.Name, err)
			return err
		}
	}
	if err := t.WM.Hide(t.ID()); err != nil {
		log.Printf("Error hiding window %s: %v", t.Name, err)
		return err
	}
	return t.SetState(NotVisible)
}

// Focus focuses the window
func (t *Tracked) Focus() error {
	log.Printf("Focusing window %s", t.Name)
	return t.WM.Focus(t.ID())
}

// IsFocused checks if the window is focused
func (t *Tracked) IsFocused() bool {
	isFocused := t.WM.IsFocused(t.ID())
	log.Printf("Checking if window %s is focused: %v", t.Name, isFocused)
	return isFocused
}

// ShowOrHide toggles the window visibility
func (t *Tracked) ShowOrHide() error {
	log.Printf("Toggling visibility for window %s", t.Name)
	previous := t.StateMgr.LoadPrevID()
	log.Printf("Previous window ID: %s", previous)

	switch t.State() {
	case Errored:
		log.Printf("Window %s is errored: ", t.Name)
		t.StateMgr.SetState(t.Name, Visible)
		return errors.New("Window is errored")
	case Visible:
		log.Printf("Window %s is visible, hiding", t.Name)
		if t.SwitchTo && !t.IsFocused() {
			log.Printf("Window %s needs focus, focusing", t.Name)
			return t.Focus()
		}
		if err := t.HideAndUpdate(); err != nil {
			log.Printf("Error hiding window %s: %v", t.Name, err)
			return err
		}
		// log.Printf("Focusing previous window %s", previous)
		// return t.WM.Focus(previous)
	case NotVisible:
		log.Printf("Window %s is not visible, showing", t.Name)
		if err := t.StateMgr.StorePrevID(t.WM.GetFocusedID()); err != nil {
			log.Printf("Error storing previous ID for window %s: %v", t.Name, err)
			return err
		}
		return t.ShowAndUpdate()
	}
	return nil
}

// ToggleAndUpdate toggles the window state and updates the state management
func (t *Tracked) ToggleAndUpdate() error {
	log.Printf("Toggling and updating window %s", t.Name)
	return t.ShowOrHide()
}
