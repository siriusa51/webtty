package apis

const (
	// User input typically from a keyboard
	Input = '1'
	// Notify that the browser size has been changed
	ResizeTerminal = '2'
	Ping           = '3'
)

const (
	Output = '1'
	Pong   = '2'
	Closed = '3'
)

type ResizeMessage struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}
