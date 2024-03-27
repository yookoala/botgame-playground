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

	// Unmarshal raw into any type
	Unmarshal(v interface{}) error

	// ReadDataTo read from the data field and write to the given type
	ReadDataTo(v interface{}) error

	// WriteDataFrom read frojm the given type and write to the data field
	WriteDataFrom(v interface{}) error
}

// Signal abstraction.
//
// For spontaneous messages that do not require a server-client
// communication (request or an event). Used for process initiation and etc.
type Signal interface {
	Message

	// Type returns the type of the message
	Type() string

	// Signal returns the signal type of the message
	Signal() string
}

// Request abstraction
type Request interface {
	Message

	// ReqeustID returns the request id of the message
	RequestID() string

	// RequestType returns the request field of the message
	RequestType() string
}

// Response abstraction
type Response interface {
	Message

	// ReqeustID returns the request id of the message
	RequestID() string

	// Response returns the response field of the message
	Response() string

	// Code returns the code field of the message
	Code() int
}

// ErrorResponse abstraction.
type ErrorResponse interface {
	Response

	// ReqeustID returns the request id of the message
	RequestID() string

	// ErrorString returns the error string of the message
	ErrorString() string
}

// Greeting message
type message struct {
	sessionID   string
	signal      string
	messageType string
	requestID   string
	requestType string
	response    string
	code        int
	data        json.RawMessage
	errorString string
	raw         []byte
}

// jsonMessage is the JSON representation of the message struct
// for read-write to and from JSON
type jsonMessage struct {
	SessionID   string          `json:"sessionID,omitempty"`
	Type        string          `json:"type,omitempty"`
	Signal      string          `json:"signal,omitempty"`
	RequestID   string          `json:"requestID,omitempty"`
	RequestType string          `json:"requestType,omitempty"`
	Response    string          `json:"response,omitempty"`
	Code        int             `json:"code,omitempty"`
	Data        json.RawMessage `json:"data,omitempty"`
	ErrorString string          `json:"error,omitempty"`
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

// Signal returns the signal type of the message
func (m *message) Signal() string {
	return m.signal
}

// Type returns the type of the message
func (m *message) Type() string {
	return m.messageType
}

// RequestID returns the request id of the message
func (m *message) RequestID() string {
	return m.requestID
}

// RequestType returns the request field of the message
func (m *message) RequestType() string {
	return m.requestType
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

// ReadDataTo read from the data field and write to the given type
func (m *message) ReadDataTo(v interface{}) error {
	return json.Unmarshal(m.data, v)
}

// WriteDataFrom marshals the given type into the data field
func (m *message) WriteDataFrom(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	m.data = b
	return nil
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
	return json.Marshal(&jsonMessage{
		SessionID:   m.sessionID,
		Signal:      m.signal,
		Type:        m.messageType,
		RequestID:   m.requestID,
		RequestType: m.requestType,
		Response:    m.response,
		Code:        m.code,
		Data:        m.data,
		ErrorString: m.errorString,
	})
}

// UnmarshalJSON unmarshals the JSON data into the message
func (m *message) UnmarshalJSON(b []byte) (err error) {
	// Unmarshal to an unnamed struct type with essential fields.
	// Return if there is an error.
	v := jsonMessage{}
	if err = json.Unmarshal(b, &v); err != nil {
		return
	}

	// Assign the fields to the message
	m.sessionID = v.SessionID
	m.signal = v.Signal
	m.messageType = v.Type
	m.requestID = v.RequestID
	m.requestType = v.RequestType
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

// NewSignal creates a new signal message
func NewSignal(signal string, data interface{}) Message {
	m := &message{
		messageType: "signal",
		signal:      signal,
	}
	if data != nil {
		b, _ := json.Marshal(data)
		m.data = b
	}
	return m
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
func NewRequest(requestID, requestType string, data interface{}) Request {
	m := &message{
		requestID:   requestID,
		requestType: requestType,
		messageType: "request",
	}
	if data != nil {
		b, _ := json.Marshal(data)
		m.data = b
	}
	return m
}

// Create a new response
func NewResponse(sessionID, requestID, requestType string, code int, response string, data interface{}) Response {
	m := &message{
		sessionID:   sessionID,
		requestID:   requestID,
		requestType: requestType,
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
