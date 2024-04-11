package comms

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sync"
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
	return NewMessageFromJSON(bytes.Trim(b, "\n"))
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

	onClose func(*Session)
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

// NewSessionFromConn creates a new Session from a newly
// dailed connection and to obtain the session ID from the
// greeting message.
func NewSessionFromConn(conn io.ReadWriteCloser) (sess *Session, greeting Message, err error) {
	greeting, err = NewMessageReader(conn).ReadMessage()
	if err != nil {
		return
	}
	sess = NewSession(greeting.SessionID(), conn)
	return
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

// OnClose sets a callback function to be called when the session is closed.
func (s *Session) OnClose(f func(*Session)) *Session {
	s.onClose = f
	return s
}

// Close closes the session
func (s *Session) Close() (err error) {
	if s.conn != nil {
		err = s.conn.Close()
		s.conn = nil
	}
	if s.onClose != nil {
		// Run the onClose callback.
		s.onClose(s)
	}
	s.onClose = nil // remove reference to callback for gc
	return
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

// SessionCollection is an abstraction of a collection of sessions.
type SessionCollection interface {
	// Has checks if a session exists in the collection.
	Has(id string) bool

	// Add adds a session to the collection.
	Add(s *Session) error

	// OnAdd registers a callback function to be called when a session is added.
	OnAdd(f func(*Session))

	// Remove removes a session from the collection.
	Remove(id string)

	// OnRemove registers a callback function to be called when a session is removed.
	OnRemove(f func(*Session))

	// Len returns the size of the collection.
	Len() int

	// Get returns a session from the collection.
	Get(id string) *Session

	// Map maps a callback to all sessions in the collection.
	Map(func(*Session))
}

// sessionCollection is the default implementation of SessionCollection
type sessionCollection struct {
	sessions map[string]*Session
	lock     *sync.RWMutex

	onAdd    []func(*Session)
	onRemove []func(*Session)

	callbackLock *sync.RWMutex
}

// NewSessionCollection creates a new session collection.
func NewSessionCollection() SessionCollection {
	return &sessionCollection{
		sessions: make(map[string]*Session),
		lock:     &sync.RWMutex{},

		onAdd:    []func(*Session){},
		onRemove: []func(*Session){},

		callbackLock: &sync.RWMutex{},
	}
}

// Has checks if a session exists in the collection.
func (sc *sessionCollection) Has(id string) bool {
	sc.lock.Lock()
	defer sc.lock.Unlock()
	_, ok := sc.sessions[id]
	return ok
}

// Add adds a session to the collection.
func (sc *sessionCollection) Add(s *Session) error {
	// Check if ID exists.
	if sc.Has(s.ID()) {
		return fmt.Errorf("session %s already exists", s.ID())
	}

	// Add session to collection.
	sc.lock.Lock()
	s.OnClose(sc.onSessionClose)
	sc.sessions[s.ID()] = s
	sc.lock.Unlock()

	// Running onAdd subscribers.
	sc.callbackLock.RLock()
	for _, f := range sc.onAdd {
		f(s)
	}
	sc.callbackLock.RUnlock()
	return nil
}

func (sc *sessionCollection) onSessionClose(s *Session) {
	sc.Remove(s.ID())
	sc.callbackLock.Lock()
	defer sc.callbackLock.Unlock()
	for _, f := range sc.onRemove {
		f(s)
	}
}

// OnAdd registers a callback function to be called when a session is added.
//
// The callback function is called with the session that is added.
// Previously added sessions does not trigger the callback registered afterwards.
func (sc *sessionCollection) OnAdd(f func(*Session)) {
	sc.callbackLock.Lock()
	sc.onAdd = append(sc.onAdd, f)
	sc.callbackLock.Unlock()
}

// OnRemove registers a callback function to be called when a session is removed.
//
// The callback function is called with the session that is removed.
// Previously removed sessions does not trigger the callback registered afterwards.
func (sc *sessionCollection) OnRemove(f func(*Session)) {
	sc.callbackLock.Lock()
	sc.onRemove = append(sc.onRemove, f)
	sc.callbackLock.Unlock()
}

// Len returns the size of the collection.
func (sc *sessionCollection) Len() (l int) {
	l = len(sc.sessions)
	return l
}

// Remove removes a session from the collection.
func (sc *sessionCollection) Remove(id string) {
	sc.lock.Lock()
	defer sc.lock.Unlock()
	delete(sc.sessions, id)
}

// Get returns a session from the collection.
func (sc *sessionCollection) Get(id string) *Session {
	sc.lock.RLock()
	defer sc.lock.RUnlock()
	s, ok := sc.sessions[id]
	if !ok {
		return nil
	}
	return s
}

// Map maps a callback to all sessions in the collection.
func (sc *sessionCollection) Map(f func(*Session)) {
	sc.lock.RLock()
	defer sc.lock.RUnlock()

	// WaitGroup to wait for all goroutines to finish.
	// Map will block until all session handlers are done.
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	for _, s := range sc.sessions {
		// Start a goroutine to handle each session.
		wg.Add(1)
		go func(s *Session) {
			f(s)
			wg.Done()
		}(s)
	}
}
