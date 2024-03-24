package comms

import (
	"encoding/json"
)

// Message abstraction
type Message interface {
	// GetSessionID returns the session ID of the message
	SessionID() string

	// GetType returns the type of the message
	Type() string

	// Unmarshal raw into any type
	Unmarshal(v interface{}) error
}

// Greeting message
type message struct {
	sessionID   string
	messageType string
	raw         []byte
}

// String returns the string representation of the message
func (m *message) String() string {
	b, _ := m.MarshalJSON()
	return string(b)
}

// GetSessionID returns the session ID of the message
func (m *message) SessionID() string {
	return m.sessionID
}

// GetType returns the type of the message
func (m *message) Type() string {
	return m.messageType
}

// Unmarshal raw into any type
func (m *message) Unmarshal(v interface{}) error {
	return json.Unmarshal(m.raw, v)
}

// MarshalJSON marshals the message into JSON data
func (m *message) MarshalJSON() ([]byte, error) {
	if m.raw != nil {
		return m.raw, nil
	}
	return json.Marshal(struct {
		SessionID string
		Type      string
	}{
		SessionID: m.sessionID,
		Type:      m.messageType,
	})
}

// UnmarshalJSON unmarshals the JSON data into the message
func (m *message) UnmarshalJSON(b []byte) (err error) {
	// Unmarshal to an unnamed struct type with essential fields.
	// Return if there is an error.
	v := struct {
		SessionID string
		Type      string
	}{}
	if err = json.Unmarshal(b, &v); err != nil {
		return
	}

	// Assign the fields to the message
	m.sessionID = v.SessionID
	m.messageType = v.Type
	m.raw = b
	return
}

// NewMessageFromJSON creates a new message from JSON data
func NewMessageFromJSON(b []byte) (Message, error) {
	m := &message{}
	if err := m.UnmarshalJSON(b); err != nil {
		return nil, err
	}
	return m, nil
}

// NewMessageFromJSONString creates a new message from JSON string
func NewMessageFromJSONString(s string) (Message, error) {
	m := &message{}
	if err := m.UnmarshalJSON([]byte(s)); err != nil {
		return nil, err
	}
	return m, nil
}

// Create a new greeting message
func NewGreeting(sessionID string) Message {
	return &message{
		sessionID:   sessionID,
		messageType: "greeting",
	}
}

func MustMessage(m Message, err error) Message {
	if err != nil {
		panic(err)
	}
	return m
}
