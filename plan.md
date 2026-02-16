# Mosugo Development Plan

**Project Name:** Mosugo  
**Description:** A minimal, Window-focused spatial notes application built with Go and Fyne. Inspired by Forma, focusing on a large zoomable canvas, strict numpad interactions, and drag-to-create card mechanics.

## 1. Technical Stack & Setup
- **Language:** Go 1.23+
- **GUI Toolkit:** Fyne v2.5.3+ (Latest Stable)
- **Target OS:** Windows (primary), Cross-platform capable
- **License:** MIT

### Recommended Project Structure
```text
/mosugo
  /cmd/mosugo/main.go        # Entry point
  /internal
    /canvas                  # Zoomable canvas logic & Dot Grid rendering
    /cards                   # Mosu (Card) widget definitions
    /tools                   # Input State Machine (Pan, Card, Draw, Erase)
    /theme                   # Custom Theme & Color Palettes
    /state                   # App state management
    /storage                 # JSON Save/Load & Calendar Index
  /assets                    # Icons, DotPattern.png
  go.mod                     # Go Module definition
  Mosugo.toml                # Fyne App Metadata
```

## 2. Core Architecture

### Input Manager (State Machine)
Interaction modes are strictly controlled via Numpad keys.

| Key | Mode | Description | Cursor |
| :--- | :--- | :--- | :--- |
| **Numpad 0** | **Pan** | Default navigation. Drag to move canvas. Click to select items. | Open Hand |
| **Numpad 1** | **Card** | Drag diagonally to create a new Mosu. | Crosshair |
| **Numpad 2** | **Draw** | Freehand drawing on the canvas. | Pencil |
| **Numpad 3** | **Erase** | Click or drag to delete strokes/items. | Eraser |

### The Canvas (Zoomable Dot Grid)
- **Visuals:** Minimalist dot grid. Dots at vertices only (no lines).
- **Implementation:** `canvas.Image` with `FillMode: canvas.ImageFillTile`.
- **Logic:** Rendering a small seamless tile (e.g., 40x40px with a single center dot). Uses coordinate offsets to handle Panning and Scale to handle Zooming within a large but finite workspace.

### Mosu (Card) Creation Logic
- **Trigger:** Active in `ToolCard` (Numpad 1).
- **Action:** Mouse Drag from `Point A` to `Point B`.
- **Result:** Calculates specific width/height based on drag distance.
- **Constraint:** Minimum size enforced (e.g., 50x50px).

## 3. Visual Design System

### Color Palette (Pre-determined)
No color pickers. Users cycle through approved sets.

**Global Palette Struct:**
```go
var Palette = struct {
    Cards []color.Color
    Ink   []color.Color
}{
    Cards: []color.Color{
        color.White,                    // Paper
        color.RGBA{255, 247, 64, 255},  // Post-It Yellow
        color.RGBA{64, 224, 208, 255},  // Turquoise
        color.RGBA{255, 105, 180, 255}, // Muted Hot Pink
    },
    Ink: []color.Color{
        color.RGBA{30, 30, 30, 255},    // Near Black
        color.RGBA{220, 50, 50, 255},   // Red
        color.RGBA{50, 50, 220, 255},   // Blue
    },
}
```

## 4. Phased Development Milestones

### Phase 1: The Zoomable Canvas
- **Goal:** A performant window with a large, scaleable dot grid.
- **Key Tasks:**
    - Initialize Fyne app `main.go`.
    - Implement `Canvas` container logic (Offset/Scale).
    - Implement Tiled Image background for the Dot Grid.
    - Wire up `Numpad 0` to toggle Pan interactions.

### Phase 2: Mosu Construction
- **Goal:** Create cards using diagonal drag.
- **Key Tasks:**
    - Implement `ToolCard` state (`Numpad 1`).
    - Handle `DragStart` -> `DragEnd` to calculate rectangle geometry.
    - Create `MosuWidget` (Custom Fyne widget with Background Rect + MultiLineEntry).
    - Implement basic card selection logic.

### Phase 3: Spatial Manipulation
- **Goal:** Move, simplify, and polish interactions.
- **Key Tasks:**
    - Enable dragging existing cards (only in `Pan Mode`).
    - Implement Snapping (Snap TopLeft to nearest 20px grid vertex).
    - Polish the "Shadow" and "Selected" visual states.

### Phase 4: Ink & Eraser
- **Goal:** Freehand drawing capabilities.
- **Key Tasks:**
    - Implement `ToolDraw` (`Numpad 2`) capturing mouse path points.
    - Render paths using `canvas.Line` segments (or `canvas.Raster` for optimization).
    - Implement `ToolErase` (`Numpad 3`) with collision detection to remove objects/strokes.

### Phase 5: Storage & Persistence
- **Goal:** Save state to disk.
- **Key Tasks:**
    - Define JSON schema for Scale, Offset, and Object List.
    - Serialize `MosuData` (ID, Content, Bounds, ColorIndex, CreatedAt).
    - Implement Load/Save dialogs.

### Phase 6: Packaging
- **Goal:** Deployable Windows Artifact.
- **Key Tasks:**
    - `fyne package -os windows`.
    - Set application icon and metadata.

## 5. Future Roadmap: The Calendar
- **Concept:** Uses `CreatedAt` timestamps from saved Mosus.
- **Interaction:** A clickable calendar view. Clicking a date lists Mosus from that day.
- **Action:** Clicking a list item "teleports" (pans) the canvas to that specific Mosu.
