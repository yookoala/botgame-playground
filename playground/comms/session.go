package comms

import (
	"bufio"
	"encoding/json"
	"io"
)

// MessageReader reads a message from an io.Reader
type MessageReader interface {
	ReadMessage() (Message, error)
}

// messageReader is the default implementation of MessageReader
type messageReader struct {
	r *bufio.Reader
}

// NewMessageReader creates a new MessageReader
func NewMessageReader(r io.Reader) MessageReader {
	return &messageReader{r: bufio.NewReader(r)}
}

// ReadMessage reads a message from the reader
func (mr *messageReader) ReadMessage() (m Message, err error) {
	b, err := mr.r.ReadBytes('\n')
	if err != nil {
		return
	}
	return NewMessageFromJSON(b)
}

// MessageWriter writes a message to an io.Writer
type MessageWriter interface {
	WriteMessage(Message) error
}

// messageWriter is the default implementation of MessageWriter
type messageWriter struct {
	encoder *json.Encoder
}

// NewMessageWriter creates a new MessageWriter
func NewMessageWriter(w io.Writer) MessageWriter {
	return &messageWriter{encoder: json.NewEncoder(w)}
}

// WriteMessage writes a message to the writer
func (mw *messageWriter) WriteMessage(m Message) error {
	return mw.encoder.Encode(m)
}

// Session represents a connection session
type Session struct {
	id   string
	conn io.ReadWriteCloser

	mr MessageReader
	mw MessageWriter
}

// NewSession creates a new Session
func NewSession(id string, conn io.ReadWriteCloser) *Session {
	return &Session{
		id:   id,
		conn: conn,
		mr:   NewMessageReader(conn),
		mw:   NewMessageWriter(conn),
	}
}

// ID returns the session ID
func (s *Session) ID() string {
	return s.id
}

// ReadMessage reads a message from the session
func (s *Session) ReadMessage() (Message, error) {
	return s.mr.ReadMessage()
}

// WriteMessage writes a message to the session
func (s *Session) WriteMessage(m Message) error {
	return s.mw.WriteMessage(m)
}

// Close closes the session
func (s *Session) Close() error {
	return s.conn.Close()
}

// SessionHandler handles a session.
type SessionHandler interface {
	HandleSession(s *Session) error
}

// SessionHandlerFunc is an adapter to allow the use of ordinary functions as SessionHandlers.
type SessionHandlerFunc func(s *Session) error

// HandleSession calls f(s)
func (f SessionHandlerFunc) HandleSession(s *Session) error {
	return f(s)
}
