package engine

// Status describes what the engine loop is currently doing.
type Status int

const (
	StatusStopped Status = iota
	StatusRunning
	StatusPaused
)

func (s Status) String() string {
	switch s {
	case StatusRunning:
		return "running"
	case StatusPaused:
		return "paused"
	default:
		return "stopped"
	}
}
