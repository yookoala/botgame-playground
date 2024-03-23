package comms

// Greeting message
type Greeting struct {
	SessionID string
	Type      string
	Message   string
}

// Create a new greeting message
func NewGreeting(sessionID string, message string) *Greeting {
	return &Greeting{
		SessionID: sessionID,
		Type:      "greeting",
		Message:   message,
	}
}
