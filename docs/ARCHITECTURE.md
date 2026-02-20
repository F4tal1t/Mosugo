# Mosugo Architecture

This document describes the high-level architecture and design decisions of Mosugo.

## Table of Contents

- [Overview](#overview)
- [Component Architecture](#component-architecture)
- [Coordinate System](#coordinate-system)
- [Tool State Machine](#tool-state-machine)
- [Persistence Model](#persistence-model)
- [Rendering Pipeline](#rendering-pipeline)
- [Performance Optimizations](#performance-optimizations)
- [Design Decisions](#design-decisions)

## Overview

Mosugo is built as a **single-window Fyne application** with a custom infinite canvas widget at its core. The architecture follows a **separation of concerns** pattern with distinct layers for:

1. **UI Layer** (`cmd/mosugo`) – Application setup, window management, tool buttons
2. **Canvas Layer** (`internal/canvas`) – Infinite canvas, coordinate transforms, rendering
3. **Widget Layer** (`internal/cards`, `internal/ui`) – Card widgets, calendar, UI components
4. **Tool Layer** (`internal/tools`) – Tool state machine, mouse event handling
5. **Persistence Layer** (`internal/storage`) – JSON-based workspace save/load
6. **Theme Layer** (`internal/theme`) – Custom colors and fonts

### Data Flow

```
User Input (Mouse/Keyboard)
    ↓
Tool (Select/Card/Draw/Erase)
    ↓
Canvas (Coordinate Transform)
    ↓
State Update (Cards, Strokes, View)
    ↓
Auto-save Timer
    ↓
Storage (JSON Persistence)
```

## Component Architecture

### Main Application (`cmd/mosugo/main.go`)

**Responsibilities**:
- Window initialization and layout
- Tool button creation and state management
- Auto-save timer setup (2-second debounce)
- Calendar integration
- Workspace loading on startup

**Key Functions**:
- `main()` – Entry point, sets up Fyne app and window
- `createToolButton()` – Creates tool buttons with icons loaded from embedded assets
- Auto-save callback triggers `canvas.SaveCurrentWorkspace()` after 2 seconds of inactivity

### Canvas (`internal/canvas`)

#### `canvas.go`

**MosugoCanvas** is the central widget that:
- Manages the infinite dot grid background
- Handles pan/zoom transformations
- Contains cards and strokes as child objects
- Delegates mouse events to the active tool
- Tracks dirty state for auto-save

**Key Methods**:
- `ScreenToWorld(pos)` / `WorldToScreen(pos)` – Coordinate transformations
- `SetTool(toolType)` – Switches active tool, updates state
- `AddCard()`, `RemoveCard()` – Card lifecycle management
- `AddStroke()` – Creates line segments for drawings
- `SaveCurrentWorkspace()` / `LoadWorkspace(date)` – Persistence interface

**State**:
- `Offset fyne.Position` – Pan offset in world coordinates
- `Scale float32` – Zoom level (1.0 = 100%)
- `ActiveTool tools.Tool` – Currently active tool instance
- `strokesMap map[*canvas.Line]StrokeCoords` – Maps line objects to world coordinates
- `currentDate time.Time` – Current workspace date

#### `grid.go`

**GridRenderer** generates the dotted grid pattern using a Fyne Raster:
- Procedurally draws dots at 30px interval (GridSize constant)
- Transforms based on current scale and offset
- Background color: `GridBg`, dots: `GridLine`

### Cards (`internal/cards/mosu.go`)

**MosuWidget** represents a note card:
- Stores content, position (world space), size, color index
- Renders with custom colored label and background
- Supports text editing via Fyne's entry widget
- Checkbox parser converts `[x]` → ☑ and `[ ]` → ☐

**Architecture**:
- Custom widget extending `widget.BaseWidget`
- Composition of `coloredLabel` for text rendering
- Manual text wrapping with `wrapText()` helper

### Tools (`internal/tools`)

#### Tool Interface

Each tool implements:
```go
type Tool interface {
    Name() string
    Cursor() desktop.Cursor
    OnTapped(c Canvas, e *fyne.PointEvent)
    OnDragged(c Canvas, e *fyne.DragEvent)
    OnDragEnd(c Canvas)
}
```

#### Tool Implementations

1. **SelectTool**: 
   - Single-click to select cards
   - Drag to move selected card
   - Right-click panning via rawPos (raw screen coordinates)

2. **CardTool**:
   - Drag gesture creates new card
   - Ghost rectangle preview during drag
   - Snaps to grid on creation

3. **DrawTool**:
   - Continuous stroke creation with `currentStroke` array
   - Stroke simplification with Douglas-Peucker (epsilon = 1.5)
   - Dual-line system: glow (4.0 width) + regular (2.5 width)

4. **EraseTool**:
   - Hover-based erasing with 12px threshold
   - Erases entire strokes (identified by strokeID)
   - Visual feedback with cursor change

#### Canvas Interface

Tools interact with canvas through the `Canvas` interface, which provides:
- Coordinate helpers (transform, snap)
- Object manipulation (add/remove)
- Stroke management (ID generation, simplification)
- Persistence (`MarkDirty()`)

This **decoupling** allows tools to be tested independently and new tools to be added easily.

### Storage (`internal/storage/storage.go`)

**Workspace Persistence Model**:

Each day gets its own JSON file: `YYYY-MM-DD.mosugo`

**Example**: `2026-02-20.mosugo`

```json
{
  "scale": 1.0,
  "offset_x": 0.0,
  "offset_y": 0.0,
  "date": "2026-02-20",
  "cards": [
    {
      "id": "uuid-goes-here",
      "content": "Sample card\n[x] Done\n[ ] Todo",
      "pos_x": 150.0,
      "pos_y": 200.0,
      "width": 180.0,
      "height": 140.0,
      "color_index": 0,
      "created_at": "2026-02-20T10:30:00Z"
    }
  ],
  "strokes": [
    {
      "p1_x": 100.0,
      "p1_y": 100.0,
      "p2_x": 200.0,
      "p2_y": 150.0,
      "color_index": 0,
      "width": 2.5,
      "stroke_id": 1
    }
  ]
}
```

**Storage Location**:
- Windows: `%APPDATA%\Roaming\Mosugo\`
- Linux: `~/.config/Mosugo/`

**Key Functions**:
- `SaveWorkspace()` – Marshals workspace state to JSON
- `LoadWorkspace()` – Loads workspace, returns empty state if file doesn't exist
- `ListSavedDates()` – Scans directory for all `.mosugo` files, parses dates

### UI Components (`internal/ui`)

#### Calendar (`calendar.go`)

**CalendarContent**:
- Month view grid with prev/next navigation
- Highlights current date and dates with saved workspaces
- Custom compact grid layout (reduced spacing)
- Callback `onDateSelected(time.Time)` when user clicks a date

#### Metaball Border (`canvas_container.go`)

**MetaballBorder**:
- Wraps canvas with animated SDF-based border
- Bottom tab for calendar (30px collapsed, 240px expanded)
- Smooth animation using `BounceEaseOut` from `tools/animation.go`
- Invisible tappable areas for interaction (no visual feedback)

**SDF Rendering**:
- Uses signed distance fields for smooth rounded rectangles
- `sdBox()` – Distance to box
- `opSmoothUnion()` – Merges shapes with smooth blending
- Renders via Fyne Raster, pixel-by-pixel evaluation

**Performance Optimization**:
- During animation, renders at lower resolution (scale factor 0.6)
- Upscales after animation completes for crisp borders

## Coordinate System

Mosugo uses **two coordinate systems**:

### World Space
- Infinite coordinate system where cards and strokes live
- Origin (0, 0) is the starting pan position
- Coordinates stored in JSON persistence

### Screen Space
- Viewport coordinates (pixels on window)
- Origin (0, 0) is top-left of canvas widget
- Used for mouse event positions

### Transformations

**World → Screen** (for rendering):
```go
func WorldToScreen(worldPos fyne.Position) fyne.Position {
    return fyne.NewPos(
        (worldPos.X + Offset.X) * Scale,
        (worldPos.Y + Offset.Y) * Scale,
    )
}
```

**Screen → World** (for mouse events):
```go
func ScreenToWorld(screenPos fyne.Position) fyne.Position {
    return fyne.NewPos(
        screenPos.X / Scale - Offset.X,
        screenPos.Y / Scale - Offset.Y,
    )
}
```

### Grid Snapping

Cards snap to 30px grid when created:
```go
func snap(v float32) float32 {
    return float32(int(v/GridSize) * GridSize)
}
```

## Tool State Machine

### State Diagram

```
[App Start] → SelectTool (default)

Numpad 1 → SelectTool
Numpad 2 → CardTool
Numpad 3 → DrawTool
Numpad 4 → EraseTool
```

### Tool Lifecycle

1. **Tool Creation**: Tools are stateful structs (e.g., `SelectTool`, `DrawTool`)
2. **Activation**: `canvas.SetTool(toolType)` creates new tool instance, sets `ActiveTool`
3. **Event Delegation**: Canvas delegates all mouse events to `ActiveTool`
4. **State Reset**: On tool switch, previous tool state is discarded

### Event Flow

```
Mouse Event (Fyne)
    ↓
MosugoCanvas.Tapped/Dragged
    ↓
ActiveTool.OnTapped/OnDragged
    ↓
Tool Logic (coordinate transform, canvas mutation)
    ↓
canvas.MarkDirty() → Triggers auto-save
    ↓
canvas.Refresh() → Redraws
```

## Persistence Model

### Auto-save Strategy

**Debounced Save** (2 seconds):
- Any canvas change calls `MarkDirty()`
- `onDirty` callback starts/restarts 2-second timer
- Timer fires → `SaveCurrentWorkspace()` writes to JSON
- Prevents excessive disk I/O during rapid drawing

### Workspace Loading

**On App Start** or **Date Selection**:
1. `LoadWorkspace(date)` called
2. Reads `YYYY-MM-DD.mosugo` from storage directory
3. If file doesn't exist, returns empty workspace (no error)
4. Canvas reconstructs cards and strokes from JSON data

### Data Integrity

- **JSON Validation**: Unmarshaling errors logged but don't crash app
- **Backward Compatibility**: Additional fields in JSON are ignored
- **No File Locking**: Single-user app assumption (no concurrent writes)

## Rendering Pipeline

### Frame Rendering Order

1. **Grid** (canvas.Raster) – Background dots
2. **Strokes** (canvas.Line) – Drawing lines (glow + regular)
3. **Cards** (cards.MosuWidget) – Note cards on top
4. **Ghost Elements** – Drag preview rectangles (only when active)

### Card Rendering

Each MosuWidget renders:
1. Background rectangle with rounded corners
2. Text content (using custom `coloredLabel`)
3. Selection highlight (blue border if selected)

### Stroke Rendering

Two-line system per stroke:
- **Glow Line**: Width 4.0, slightly transparent, InkLightGrey
- **Regular Line**: Width 2.5, opaque, InkGrey

Both lines stored in `strokesMap` with world coordinates.

### Metaball Border

Rendered as a Fyne Raster with custom pixel shader:
- For each pixel, calculate distance to all UI elements (tabs, etc.)
- Apply `opSmoothUnion` to blend distances
- Color pixel if distance < threshold

## Performance Optimizations

### 1. Douglas-Peucker Stroke Simplification

**Problem**: Raw mouse tracking creates hundreds of line segments per stroke

**Solution**: Simplify strokes with epsilon = 1.5 after drawing
```go
func SimplifyStroke(points []fyne.Position, epsilon float32) []fyne.Position
```

**Impact**: Reduces stroke complexity by 50-80%, improves rendering and JSON size

### 2. Metaball Resolution Scaling

**Problem**: SDF evaluation is computationally expensive (O(pixels × shapes))

**Solution**: During animation, render at 60% resolution, upscale
```go
if m.animating {
    scaleFactor = 0.6  // Lower resolution
} else {
    scaleFactor = 1.0  // Full resolution
}
```

**Impact**: Maintains 60 FPS during tab animation

### 3. Grid Pattern Caching

Grid is rendered once as a Raster, only regenerates on scale/offset change.

### 4. Lazy Calendar Rendering

Calendar widget only created when tab expanded, destroyed on collapse.

## Design Decisions

### Why Fyne?

- **Cross-platform** with native feel (via OpenGL)
- **Single binary** executable (no runtime dependencies)
- **Go-native** – No C++ bindings or complex build setup
- **Custom widgets** easy to implement

### Why Daily Workspaces?

**Rationale**: Infinite canvas for each day provides:
- Temporal organization without manual folder management
- Easy navigation through calendar
- Reduced file size per workspace

**Alternative Considered**: Single infinite canvas with date tags
- **Rejected**: Would require search/filter UI, harder to implement

### Why JSON over Binary Format?

- **Human-readable**: Easy to debug, inspect, or recover data
- **Version control friendly**: Text diffs work
- **Simple marshaling**: Native Go `encoding/json` support

**Trade-off**: Larger file size (acceptable for typical workspaces)

### Why Embedded Assets?

- **Single executable**: No "assets" folder to distribute
- **No path issues**: Relative paths break when run from different CWD
- **Professional deployment**: Users just download .exe

Implemented with Go 1.16+ `embed` package.

### Why No Undo/Redo in v1.0?

**Complexity vs. Value**:
- Requires command pattern or state snapshots
- JSON serialization overhead for each action
- Auto-save makes undo less critical ("reload yesterday's workspace")

**Future**: May add with in-memory undo stack (before auto-save)

### Why Metaball Borders?

**Aesthetic Choice**: Creates a distinct, "hand-drawn" feel
- Inspired by Forma's organic UI
- SDF technique learned from shader programming
- Tab animation provides smooth UX

**Trade-off**: Higher render cost (mitigated by resolution scaling)

---

## Future Architecture Considerations

### Multi-User Sync
If collaboration is added:
- Need file locking or operational transforms
- Consider moving to database (SQLite?)
- Add workspace versioning

### Plugin System
Could expose `Tool` interface for third-party tools:
- Load tools from DLLs/shared libraries
- Register with tool dispatcher

### Performance at Scale
For workspaces with 1000+ cards:
- Implement spatial indexing (quadtree)
- Cull off-screen objects from rendering
- Lazy-load strokes

### Mobile Support
Fyne supports iOS/Android:
- Touch gestures for pan/zoom
- Virtual keyboard for card editing
- Storage to platform-specific directories

---

## Contributing to Architecture

When proposing architectural changes:
1. Open an issue describing the problem and proposed solution
2. Consider backward compatibility (can old workspaces still load?)
3. Update this document with your changes
4. Add tests for new components

For questions about architecture decisions, see [CONTRIBUTING.md](../CONTRIBUTING.md) or open a discussion.
