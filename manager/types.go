package manager

// WindowState represents the visibility state of a window
type WindowState int

const (
	Errored WindowState = iota
	Visible
	NotVisible
)

// WindowType represents the type of window being tracked
type WindowType int

const (
	TypeFocused WindowType = iota
	TypeApplication
	TypeClean
	TypeHide
	TypeHideLatest
	TypeHideAll
	TypeShowAll
)

// StateManagement defines the interface for state persistence
type StateManagement interface {
	GetID(name string) string
	StoreID(name, id string) error
	DestroyID(name string) error
	SetState(name string, state WindowState) error
	LatestShown(name string) (string, error)
	LatestCount() int
	IsLatestEmpty() bool
	RemoveFromLatest(name string) error
	GetState(id string) WindowState
	IsTracked(name string) bool
	SaveCurrent(name string, windowType WindowType, focusedID string) error
	StorePrevID(id string) error
	LoadPrevID() string
	AllHidden() []struct {
		Name string
		ID   string
	}
	ResetAll() error
	AllTracked() map[string]string
}
