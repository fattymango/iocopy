package executor

import "copy/internal/model"

// InputExecutor executes keyboard and mouse input on the local system
type InputExecutor interface {
	ExecuteKeyboard(event model.KeyboardEvent) error
	ExecuteMouseMove(event model.MouseMoveEvent) error
	ExecuteMouseClick(event model.MouseClickEvent) error
	ExecuteMouseScroll(event model.MouseScrollEvent) error
	Close() error
}
