package comms_test

import (
	"testing"

	"github.com/yookoala/botgame-playground/playground/comms"
)

func TestNewGreeting(t *testing.T) {
	// Create a new greeting message
	m := comms.NewGreeting("123")

	// Check the session ID
	if expected, actual := "123", m.SessionID(); expected != actual {
		t.Errorf("session ID is not correct. expected %#v, got %#v", expected, actual)
	}

	// Check the message type
	if expected, actual := "greeting", m.Type(); expected != actual {
		t.Errorf("type is not correct. expected %#v, got %#v", expected, actual)
	}
}
