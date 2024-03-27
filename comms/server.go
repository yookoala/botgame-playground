package comms

import (
	"context"
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

// getNewSessionIDs allocate a new session ID for use.
func getNewSessionIDs() <-chan string {
	ch := make(chan string)
	go func() {
		maxInt := int(^uint(0) >> 1)
		for id := 1; true; id++ {
			hash := sha1.New()
			hash.Write([]byte(fmt.Sprintf("%d.%d", id, time.Now().UnixMicro())))
			hashId := hash.Sum(nil)
			ch <- fmt.Sprintf("%x", hashId)[0:12]
			if id == maxInt {
				id = 0
			}
		}
	}()
	return ch
}

// StartServer creates a new server loop and start listening to the listener.
func StartServer(listener net.Listener, sh SessionHandler) (err error) {
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
				return nil
			default:
				log.Printf("Socket error: %v", err)
				return err
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
	sc SessionCollection
	mq chan ContextMessage

	lock    *sync.RWMutex
	stopped bool
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
		go func(smq *SimpleMessageQueue, s *Session) {
			for {
				// Try reading message from the session.
				m, err := s.ReadMessage()
				if err == io.EOF {
					// Session closed. Terminate reading loop.
					log.Printf("SimpleMessageQueue session closed: %s", s.ID())
					s.Close()
					return
				} else if err != nil {
					// Unexpected error in reading message. Log and terminate reading loop.
					return
				}

				// Fan-in messages to a single queue.
				smq.Enqueue(
					WithSessionID(ctx, s.ID()),
					m,
				)
			}
		}(smq, s)
	})

	// Note: intentionally blocking before onAdd is correctly called.
	//       do not run Start() in another goroutine or message might
	//       arrive before the queue is ready. In that case, message
	//       will be lost.

	go func(smq *SimpleMessageQueue, mh MessageHandler, mw MessageWriter) {
		for {
			ctx, m, err := smq.Dequeue()
			if err == io.EOF {
				// Queue stopped,
				return
			} else if err != nil {
				panic(err)
			}

			mh.HandleMessage(ctx, m, mw)
		}
	}(smq, mh, mw)
}

// Enqueue sends a message to the message queue.
func (smq *SimpleMessageQueue) Enqueue(ctx context.Context, m Message) (err error) {

	// Enqueue should check if the queue is stopped first
	// because sending to closed channel will panic.
	// There is no way to detect channel closed without
	// reading from, which Enqueue should not do.
	smq.lock.RLock()
	defer smq.lock.RUnlock()
	if smq.stopped {
		// If the queue is stopped, terminate the reading loop for the session.
		return io.EOF
	}

	smq.mq <- ContextMessage{
		Context: ctx,
		Message: m,
	}
	return
}

// Dequeue receives a message from the message queue.
func (smq *SimpleMessageQueue) Dequeue() (ctx context.Context, m Message, err error) {
	// Dequeue won't lock because it is easy to check if the read-channel is closed.
	cm, ok := <-smq.mq
	if !ok {
		return nil, nil, io.EOF
	}
	return cm.Context, cm.Message, nil
}

// Stop stops the message queue.
func (smq *SimpleMessageQueue) Stop() {

	// check if the queue is already stopped / stopping by another
	// goroutine.
	smq.lock.RLock()
	stopped := smq.stopped
	smq.lock.RUnlock()
	if stopped {
		// already stopping or stopped. Do nothing.
		return
	}

	// declare the queue stopped and close the channel.
	smq.lock.Lock()
	smq.stopped = true
	close(smq.mq)
	smq.lock.Unlock()
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
		sc: sc,
		mq: mq,

		lock:    &sync.RWMutex{},
		stopped: false,
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
