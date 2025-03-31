package wm

import (
	"fmt"

	"github.com/hellola/startorswitch/config"
)

// Factory creates window manager implementations
type Factory struct{}

// NewFactory creates a new window manager factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateWM creates a window manager implementation based on configuration
func (f *Factory) CreateWM(cfg *config.Config) (WMIntegration, error) {
	switch cfg.WindowManager {
	case "bspwm":
		return NewBSPWMIntegration(), nil
	case "i3":
		return NewI3Integration(), nil
	default:
		return nil, fmt.Errorf("unsupported window manager: %s", cfg.WindowManager)
	}
}
