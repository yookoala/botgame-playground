package comms

import (
	"encoding/json"
	"fmt"
)

// Message abstraction
type Message interface {
	// SessionID returns the session ID of the message
	SessionID() string

	// Type returns the type of the message
	Type() string

	// UnmarshalData unmarshals the data field into the given type
	UnmarshalData(v interface{}) error

	// Unmarshal raw into any type
	Unmarshal(v interface{}) error
}

// Request abstraction
type Request interface {
	Message

	// Request returns the request field of the message
	Request() string
}

// Response abstraction
type Response interface {
	Message

	// Response returns the response field of the message
	Response() string

	// Code returns the code field of the message
	Code() int
}

// ErrorResponse abstraction.
type ErrorResponse interface {
	Response

	ErrorString() string
}

// Greeting message
type message struct {
	sessionID   string
	messageType string
	requestID   string
	request     string
	response    string
	code        int
	data        json.RawMessage
	errorString string
	raw         []byte
}

// String returns the string representation of the message
func (m *message) String() string {
	b, _ := m.MarshalJSON()
	return string(b)
}

// SessionID returns the session ID of the message
func (m *message) SessionID() string {
	return m.sessionID
}

// Type returns the type of the message
func (m *message) Type() string {
	return m.messageType
}

// Request returns the request field of the message
func (m *message) Request() string {
	return m.request
}

// Response returns the response field of the message
func (m *message) Response() string {
	return m.response
}

// Code returns the code field of the message
func (m *message) Code() int {
	return m.code
}

// Error returns the error string of the message
func (m *message) ErrorString() string {
	return m.errorString
}

// UnmarshalData unmarshals the data into the given type
func (m *message) UnmarshalData(v interface{}) error {
	return json.Unmarshal(m.data, v)
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
		SessionID string          `json:"sessionID,omitempty"`
		Type      string          `json:"type,omitempty"`
		RequestID string          `json:"requestID,omitempty"`
		Request   string          `json:"request,omitempty"`
		Response  string          `json:"response,omitempty"`
		Code      int             `json:"code,omitempty"`
		Data      json.RawMessage `json:"data,omitempty"`
		Error     string          `json:"error,omitempty"`
	}{
		SessionID: m.sessionID,
		Type:      m.messageType,
		RequestID: m.requestID,
		Request:   m.request,
		Response:  m.response,
		Code:      m.code,
		Data:      m.data,
		Error:     m.errorString,
	})
}

// UnmarshalJSON unmarshals the JSON data into the message
func (m *message) UnmarshalJSON(b []byte) (err error) {
	// Unmarshal to an unnamed struct type with essential fields.
	// Return if there is an error.
	v := struct {
		SessionID   string          `json:"sessionID,omitempty"`
		Type        string          `json:"type,omitempty"`
		RequestID   string          `json:"requestID,omitempty"`
		Request     string          `json:"request,omitempty"`
		Response    string          `json:"response,omitempty"`
		Code        int             `json:"code,omitempty"`
		Data        json.RawMessage `json:"data,omitempty"`
		ErrorString string          `json:"error,omitempty"`
	}{}
	if err = json.Unmarshal(b, &v); err != nil {
		return
	}

	// Assign the fields to the message
	m.sessionID = v.SessionID
	m.messageType = v.Type
	m.requestID = v.RequestID
	m.request = v.Request
	m.response = v.Response
	m.code = v.Code
	m.data = v.Data
	m.errorString = v.ErrorString
	m.raw = b
	return
}

// NewMessageFromJSON creates a new message from JSON data
func NewMessageFromJSON(b []byte) (Message, error) {
	m := &message{}
	if err := m.UnmarshalJSON(b); err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON string: %s, JSON: %s", err, string(b))
	}
	return m, nil
}

// NewMessageFromJSONString creates a new message from JSON string
func NewMessageFromJSONString(s string) (Message, error) {
	m := &message{}
	if err := m.UnmarshalJSON([]byte(s)); err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON string: %s, JSON: %s", err, s)
	}
	return m, nil
}

// Create a new simple message
func NewSimpleMessage(sessionID, messageType string) Message {
	return &message{
		sessionID:   sessionID,
		messageType: messageType,
	}
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

// Create a new request
func NewRequest(requestID string, data interface{}) Request {
	m := &message{
		requestID:   requestID,
		messageType: "request",
	}
	if data != nil {
		b, _ := json.Marshal(data)
		m.data = b
	}
	return m
}

// Create a new response
func NewResponse(sessionID, requestID string, code int, response string, data interface{}) Response {
	m := &message{
		sessionID:   sessionID,
		requestID:   requestID,
		code:        code,
		response:    response,
		messageType: "response",
	}
	if data != nil {
		b, _ := json.Marshal(data)
		m.data = b
	}
	return m
}

// NewErrorResposne creates a new error response
func NewErrorResponse(sessionID, requestID string, code int, response, errorString string) ErrorResponse {
	return &message{
		sessionID:   sessionID,
		requestID:   requestID,
		code:        code,
		response:    response,
		errorString: errorString,
		messageType: "response",
	}
}
