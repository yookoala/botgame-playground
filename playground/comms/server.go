package comms

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

// getNewSessionIDs allocate a new session ID for use.
func getNewSessionIDs() <-chan string {
	ch := make(chan string)
	go func() {
		maxInt := int(^uint(0) >> 1)
		for id := 1; true; id++ {
			ch <- fmt.Sprintf("%d", id)
			if id == maxInt {
				id = 0
			}
		}
	}()
	return ch
}

// StartServer creates a new server loop and start listening to the listener.
func StartServer(listener net.Listener, sh SessionHandler) {
	defer listener.Close()

	newSessionIDs := getNewSessionIDs()
	for {
		sessionID := <-newSessionIDs
		conn, err := listener.Accept()
		if err != nil {
			switch err.(type) {
			case *net.OpError:
				log.Print("Socket closed. Quit")
				// TODO: some clean up to existing sessions?
				os.Exit(0)
			default:
				log.Printf("Socket error: %v", err)
				// TODO: some clean up to existing sessions?
				os.Exit(1)
			}
		}

		go sh.HandleSession(NewSession(sessionID, conn))
	}
}

// MessageWriter represents a writer that can write a message.
type MessageHandler interface {
	HandleMessage(m Message, out MessageWriter) error
}

// MessageWriter represents a writer that can write a message.
type MessageHandlerFunc func(m Message, out MessageWriter) error

// HandleMessage calls the underlying function.
// Implements MessageHandler interface.
func (f MessageHandlerFunc) HandleMessage(m Message, out MessageWriter) error {
	return f(m, out)
}

// SimpleMessageBroker helps route / multicast Message to different
// sessions. Implements MessageWriter interface.
type SimpleMessageBroker struct {
	sessions *SessionCollection
}

// NewSimpleMessageBroker creates a new SimpleMessageRouter
func NewSimpleMessageBroker(sessions *SessionCollection) *SimpleMessageBroker {
	return &SimpleMessageBroker{
		sessions: sessions,
	}
}

// WriteMessage handles the message by writing it to the appropriate session based
// on the message type.
//
// Response are sent to the specified session id. Events are broadcasted to all
// sessions.
func (r *SimpleMessageBroker) WriteMessage(m Message) error {
	if m.Type() == "response" {
		sess := r.sessions.Get(m.SessionID())
		if sess == nil {
			return fmt.Errorf("session %s not found", m.SessionID())
		}
		sess.WriteMessage(m)
		return nil
	}
	if m.Type() == "event" {
		errs := NewRouterErrorCollection()
		for _, sess := range r.sessions.sessions {
			err := sess.WriteMessage(m)
			if err != nil {
				errs.Add(err)
			}
		}
		if errs.Len() > 0 {
			return errs
		}
		return nil
	}

	return fmt.Errorf("unsupported message type: %s", m.Type())
}

// RouterErrorCollection is a collection of errors. Implements error interface.
type RouterErrorCollection struct {
	errors []error
}

// NewRouterErrorCollection creates a new RouterErrorCollection
func NewRouterErrorCollection() *RouterErrorCollection {
	return &RouterErrorCollection{
		errors: make([]error, 0),
	}
}

// Add adds an error to the collection.
func (rec *RouterErrorCollection) Add(err error) {
	rec.errors = append(rec.errors, err)
}

// Len returns the number of errors in the collection.
func (rec *RouterErrorCollection) Len() int {
	return len(rec.errors)
}

// Errors returns the errors in the collection.
func (rec *RouterErrorCollection) Errors() []error {
	return rec.errors
}

// Error returns the string representation of the errors in the collection.
// Returns empty string if there are no errors here.
func (rec *RouterErrorCollection) Error() string {
	if len(rec.errors) == 0 {
		return ""
	}

	out := ""
	for _, err := range rec.errors {
		out += err.Error() + "\n"
	}
	return fmt.Sprintf("router errors: %v", strings.TrimRight(out, "\n"))
}
