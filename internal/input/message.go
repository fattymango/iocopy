package input

// InputEvent represents a keyboard or mouse input event
type InputEvent struct {
	Type string `json:"type"` // "keyboard", "mouse_move", "mouse_click", "mouse_scroll"
	Data string `json:"data"` // JSON-encoded event data
}

// KeyboardEvent represents a keyboard key event
type KeyboardEvent struct {
	Key     string `json:"key"`      // Key name (e.g., "a", "Enter", "Ctrl")
	Action  string `json:"action"`   // "press", "release"
	Modifiers []string `json:"modifiers"` // Modifier keys: "ctrl", "shift", "alt", "meta"
}

// MouseMoveEvent represents mouse movement
type MouseMoveEvent struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// MouseClickEvent represents a mouse button click
type MouseClickEvent struct {
	Button string `json:"button"` // "left", "right", "middle"
	Action string `json:"action"` // "press", "release"
	X      int    `json:"x"`
	Y      int    `json:"y"`
}

// MouseScrollEvent represents mouse wheel scroll
type MouseScrollEvent struct {
	DeltaX int `json:"delta_x"`
	DeltaY int `json:"delta_y"`
}


