package comms_test

import (
	"encoding/json"
	"testing"

	"github.com/yookoala/botgame-playground/playground/comms"
)

func TestNewSimpleMessage(t *testing.T) {
	// Create a new simple message
	m := comms.NewSimpleMessage("123", "test")

	// Check the session ID
	if expected, actual := "123", m.SessionID(); expected != actual {
		t.Errorf("session ID is not correct. expected %#v, got %#v", expected, actual)
	}

	// Check the message type
	if expected, actual := "test", m.Type(); expected != actual {
		t.Errorf("type is not correct. expected %#v, got %#v", expected, actual)
	}

	// Unmarshal into a map and check the JSON values
	b, _ := json.Marshal(m)
	tm := make(map[string]interface{})
	json.Unmarshal(b, &tm)
	if expected, actual := "123", tm["SessionID"]; expected != actual {
		t.Errorf("session ID is not correct. expected %#v, got %#v", expected, actual)
	}
	if expected, actual := "test", tm["Type"]; expected != actual {
		t.Errorf("type is not correct. expected %#v, got %#v", expected, actual)
	}
}

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
