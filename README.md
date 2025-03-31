# StartOrSwitch

A Go utility for managing window visibility in bspwm and i3, inspired by the original Ruby implementation.

## Features

- Track and manage window visibility
- Support for focused windows and applications
- Redis-based state management
- Command-line interface for window operations
- Support for multiple window managers (bspwm, i3)

## Commands

- `f <name>` - Track focused window
- `a <name>` - Track application window
- `c <name>` - Clean (remove) tracked window
- `h` - Hide currently focused tracked window
- `hl` - Toggle latest window
- `ha` - Hide all tracked windows
- `s` - Show all hidden windows
- `r` - Reset all tracking

## Options

- `switch_to` - Switch to window when showing
- `top_padding=<value>` - Set top padding when hiding
- `mods=sticky` - Make window sticky

## Requirements

- Go 1.21 or later
- Redis server
- One of the supported window managers:
  - bspwm
  - i3
- xdotool

## Configuration

Create a configuration file at `~/.config/startorswitch/config.json`:

```json
{
  "window_manager": "bspwm",  // or "i3"
  "redis_addr": "localhost:6379"
}
```

## Installation

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod download
   ```
3. Build the binary:
   ```bash
   go build
   ```

## Usage

```bash
# Track focused window
./startorswitch f mywindow

# Track application window
./startorswitch a myapp

# Hide currently focused tracked window
./startorswitch h

# Toggle latest window
./startorswitch hl

# Hide all tracked windows
./startorswitch ha

# Show all hidden windows
./startorswitch s

# Reset all tracking
./startorswitch r
```

## Window Manager Support

### bspwm
- Uses bspwm's native commands for window management
- Supports window hiding using bspwm's hidden flag
- Supports sticky windows

### i3
- Uses i3-msg for window management
- Implements window hiding using i3's scratchpad feature
- Supports window focusing and movement

## License

MIT License



