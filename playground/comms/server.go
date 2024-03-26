package comms

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
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

// StartService creates a new server loop and start listening to the listener.
func StartService(listener net.Listener, sh SessionHandler) {
	defer listener.Close()

	log.Printf("start listening on %s", listener.Addr().String())
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

		sess := NewSession(sessionID, conn)
		log.Printf("received new session to handle: %s", sess.ID())
		go sh.HandleSession(sess)
	}
}

// MessageWriter represents a writer that can write a message.
type MessageHandler interface {
	HandleMessage(ctx context.Context, m Message, out MessageWriter) error
}

// MessageWriter represents a writer that can write a message.
type MessageHandlerFunc func(ctx context.Context, m Message, out MessageWriter) error

// HandleMessage calls the underlying function.
// Implements MessageHandler interface.
func (f MessageHandlerFunc) HandleMessage(ctx context.Context, m Message, out MessageWriter) error {
	return f(ctx, m, out)
}

type ContextMessage struct {
	Context context.Context
	Message Message
}

// SimpleMessagegQueue is a simple message queue that can be used to
// collect all messages from a session collection, then send them to
// a message handler.
//
// To simplify server implementations. Messages are collected and send
// to the MessageHandler in a linear manner.
type SimpleMessageQueue struct {
	sc   SessionCollection
	mq   chan ContextMessage
	term chan bool
}

// Start starts the message queue and start sending messages to the
// message writer.
//
// The start function will block until the message queue is ready to
// accept new session. Please do not run Start in a goroutine in parallel
// to adding session.
func (smq *SimpleMessageQueue) Start(mh MessageHandler, mw MessageWriter) {

	// Create a new context with the session collection.
	ctx := WithSessionCollection(context.Background(), smq.sc)

	// When every session is added to the collection, start a goroutine
	// that reads messages from the session and send them to the message.
	//
	// Terminate the goroutine when either:
	// 1. The reader of the session is closed (EOF); or
	// 2. The message queue is stopped
	smq.sc.OnAdd(func(s *Session) {
		log.Printf("SimpleMessageQueue get session added: %s", s.ID())
		go func(mq chan<- ContextMessage, term <-chan bool) {
			for {
				select {
				case <-smq.term:
					// Terminate the reading loop.
					return
				default:
					//log.Printf("SimpleMessageQueue pending message from session: %s", s.ID())
					m, err := s.ReadMessage()
					if err == io.EOF {
						// Session closed. Terminate reading loop.
						log.Printf("SimpleMessageQueue session closed: %s", s.ID())
						s.Close()
						return
					}
					if err != nil {
						// Unexpected error in reading message. Log and terminate reading loop.
						return
					}
					//log.Printf("SimpleMessageQueue got message from session: %s, %s", s.ID(), m)

					// Fan-in messages to a single queue.
					mq <- ContextMessage{
						Context: WithSessionID(ctx, s.ID()),
						Message: m,
					}
				}
			}
		}(smq.mq, smq.term)
	})

	// Note: intentionally blocking before onAdd is correctly called.
	//       do not run Start() in another goroutine or message might
	//       arrive before the queue is ready. In that case, message
	//       will be lost.

	go func(mq <-chan ContextMessage, term <-chan bool, mh MessageHandler, mw MessageWriter) {
		for {
			select {
			case <-smq.term:
				// Terminate the message queue.
				return
			case cm := <-smq.mq:
				mh.HandleMessage(cm.Context, cm.Message, mw)
			}
		}
	}(smq.mq, smq.term, mh, mw)
}

// Stop stops the message queue.
func (smq *SimpleMessageQueue) Stop() {
	select {
	case <-smq.term:
		// Already stopping or stopped. Do nothing.
		return
	default:
		log.Printf("Stopping SimpleMessageQueue...")
		close(smq.term)
		close(smq.mq)
	}
}

// HandleSession adds a session to the message queue.
//
// Implements SessionHandler interface.
func (smq *SimpleMessageQueue) HandleSession(s *Session) error {
	log.Printf("SimpleMessageQueue add session to queue: %s", s.ID())
	return smq.sc.Add(s)
}

// NewSimpleMessageQueue creates a new SimpleMessageQueue.
//
// This is for game server to fan-in incoming messages from
// multiple sessions into a single queue.
//
// bufferSize specify the size of the buffer for the message queue.
// small buffer will block reading from client. A non-zero positive number
// in buffer will allow client messages to read through before previous
// messages are processed.
func NewSimpleMessageQueue(sc SessionCollection, bufferSize int) *SimpleMessageQueue {
	mq := make(chan ContextMessage, bufferSize)
	return &SimpleMessageQueue{
		sc:   sc,
		mq:   mq,
		term: make(chan bool),
	}
}

// SimpleMessageBroker helps route / multicast Message to different
// sessions. Implements MessageWriter interface.
type SimpleMessageBroker struct {
	sessions SessionCollection
}

// NewSimpleMessageBroker creates a new SimpleMessageRouter
//
// This is for game server to distribute outgoing messages to
// different sessions.
func NewSimpleMessageBroker(sessions SessionCollection) *SimpleMessageBroker {
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
	//log.Printf("SimpleMessageBroker prepare to broke message: %s", m)
	if m.Type() == "response" {
		// log.Printf("SimpleMessageBroker will send response to session: %s, message: %s", m.SessionID(), m)
		sess := r.sessions.Get(m.SessionID())
		if sess == nil {
			//log.Printf("SimpleMessageBroker session not found: %s", m.SessionID())
			return fmt.Errorf("session %s not found", m.SessionID())
		}
		sess.WriteMessage(m)
		return nil
	}
	if m.Type() == "event" {
		log.Printf("SimpleMessageBroker will broadcast event to all sessions (len=%d), message: %s", r.sessions.Len(), m)
		errs := NewRouterErrorCollection()
		r.sessions.Map(func(sess *Session) {
			log.Printf("SimpleMessageBroker: send to session %s, message: %s", sess.ID(), m)
			err := sess.WriteMessage(m)
			if err != nil {
				errs.Add(err)
			}
		})
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
	lock   *sync.Mutex
}

// NewRouterErrorCollection creates a new RouterErrorCollection
func NewRouterErrorCollection() *RouterErrorCollection {
	return &RouterErrorCollection{
		errors: make([]error, 0),
		lock:   &sync.Mutex{},
	}
}

// Add adds an error to the collection.
func (rec *RouterErrorCollection) Add(err error) {
	rec.lock.Lock()
	defer rec.lock.Unlock()
	rec.errors = append(rec.errors, err)
}

// Len returns the number of errors in the collection.
func (rec *RouterErrorCollection) Len() int {
	rec.lock.Lock()
	defer rec.lock.Unlock()
	return len(rec.errors)
}

// Errors returns the errors in the collection.
func (rec *RouterErrorCollection) Errors() []error {
	rec.lock.Lock()
	defer rec.lock.Unlock()
	return rec.errors
}

// Error returns the string representation of the errors in the collection.
// Returns empty string if there are no errors here.
func (rec *RouterErrorCollection) Error() string {
	if len(rec.errors) == 0 {
		return ""
	}

	rec.lock.Lock()
	defer rec.lock.Unlock()
	out := ""
	for _, err := range rec.errors {
		out += err.Error() + "\n"
	}
	return fmt.Sprintf("router errors: %v", strings.TrimRight(out, "\n"))
}
