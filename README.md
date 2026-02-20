# Mosugo

**Mosugo** is a minimal, spatial notes application built for Windows with Go and Fyne. It combines the infinite creative freedom of a zoomable canvas with the structured simplicity of daily workspaces.

Create cards, draw freehand strokes, and organize your thoughts in an infinite space — with everything automatically saved per day.

## Features

- **Infinite Zoomable Canvas**: Pan and zoom across an unlimited dotted grid workspace
- **Four Tool Modes**: Select, Card, Draw, and Erase (switch via numpad 1-4)
- **Card System**: Create draggable note cards with markdown-like checkbox syntax `[x]` and `[]`
- **Freehand Drawing**: Smooth drawing with automatic stroke simplification (Douglas-Peucker algorithm)
- **Daily Workspaces**: Each day gets its own workspace file with automatic persistence
- **Calendar Navigation**: Quickly jump between dates to review past workspaces  
- **Auto-save**: Changes are automatically saved after 2 seconds of inactivity
- **Custom Theme**: Beautiful color palette with Comic Sans font for a friendly feel

## Installation

### Pre-built Binary (Recommended)

Download the latest `Mosugo.exe` from the [Releases](https://github.com/F4tal1t/Mosugo/releases) page.

No installation required — just download and run!

### Build from Source

**Requirements:**
- Go 1.23 or later
- Windows
- C compiler (gcc/mingw64 for Windows)

**Steps:**

```powershell
# Clone the repository
git clone https://github.com/F4tal1t/Mosugo.git
cd Mosugo

# Build the executable
go build -o mosugo.exe cmd/mosugo/main.go

# Or use Fyne packaging for a bundled executable
go install fyne.io/fyne/v2/cmd/fyne@latest
fyne package -os windows

# Run
./mosugo.exe
```

## Usage

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| **Numpad 1** | Select Tool (move and select cards) |
| **Numpad 2** | Card Tool (create new cards) |
| **Numpad 3** | Draw Tool (freehand drawing) |
| **Numpad 4** | Erase Tool (remove strokes) |

### Mouse Controls

- **Pan**: Hold middle mouse button and drag (or use right-click in Select mode)
- **Zoom**: Scroll mouse wheel
- **Select Tool**: Click cards to select, drag to move
- **Card Tool**: Drag to create a new card, type to edit
- **Draw Tool**: Click and drag to draw
- **Erase Tool**: Hover over strokes to remove them

### Card Syntax

Cards support simple markdown-like checkboxes:

- `[ ]` renders as an unchecked box: ☐
- `[x]` renders as a checked box: ☑

Example:
```
Shopping List
[ ] Milk
[x] Eggs
[ ] Bread
```

### Calendar

Click the date indicator at the bottom of the screen to open the calendar. Navigate between months and select any date to load that day's workspace.

## Data Storage

Workspaces are saved as JSON files in:

- **Windows**: `%APPDATA%\Roaming\Mosugo\`
- **Linux**: `~/.config/Mosugo/`

Each file is named `YYYY-MM-DD.mosugo` (e.g., `2026-02-20.mosugo`).

Files contain:
- All cards (content, position, size, color)
- All drawing strokes  
- Canvas view state (zoom level, pan offset)

You can back up or transfer your data by copying these files.

## Development

### Project Structure

```
Mosugo/
├── cmd/
│   └── mosugo/        # Main application entry point
├── internal/
│   ├── canvas/        # Infinite canvas and coordinate transforms
│   ├── cards/         # Card widget implementation
│   ├── storage/       # Workspace persistence layer
│   ├── theme/         # Custom Fyne theme
│   ├── tools/         # Tool state machine (Select/Card/Draw/Erase)
│   └── ui/            # Calendar and metaball border UI
├── assets/            # Icons, fonts, resources (embedded at build)
├── go.mod             # Go module definition
└── Mosugo.toml        # Fyne packaging configuration
```

### Building

```powershell
# Run in development mode
go run cmd/mosugo/main.go

# Build for production
go build -o mosugo.exe cmd/mosugo/main.go

# Package with Fyne (includes icon and metadata)
fyne package -os windows
```

### Testing

```powershell
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

### Linting

```powershell
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linters
golangci-lint run
```

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on:
- Setting up your development environment  
- Code style and standards
- Submitting pull requests
- Finding good first issues

## License

Mosugo is licensed under the [MIT License](LICENSE).

Copyright © 2026 Dibyendu Sahoo

## Acknowledgments

- Built with [Fyne](https://fyne.io/) UI framework
- Inspired by [Forma](https://www.forma.app/) and [Tweek](https://tweek.so/)
- Uses Douglas-Peucker algorithm for stroke simplification
- Metaball borders powered by Signed Distance Fields

## Support

- **Issues**: Report bugs or request features via [GitHub Issues](https://github.com/F4tal1t/Mosugo/issues)

---

Made with Pain by [Dibyendu Sahoo](https://github.com/F4tal1t)
