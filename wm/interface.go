package wm

// WMIntegration defines the interface for window manager operations
type WMIntegration interface {
	Show(nodeID string) error
	Hide(nodeID string) error
	StillAlive(nodeID string) bool
	Focus(nodeID string) error
	IsFocused(nodeID string) bool
	GetFocusedID() string
	FindOrStartApplication(name string) (string, error)
}
