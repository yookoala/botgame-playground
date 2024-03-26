package comms_test

import (
	"context"
	"io"
	"sync"
	"testing"

	"github.com/yookoala/botgame-playground/playground/comms"
)

type dummyConn struct {
	in      chan<- []byte
	out     <-chan []byte
	term    chan bool
	partner io.ReadWriteCloser
}

func NewDummyConns(size int) (serverConn, clientConn io.ReadWriteCloser) {
	chanServerToClient := make(chan []byte, size)
	chanClientToServer := make(chan []byte, size)
	term := make(chan bool)

	serverDummyConn := &dummyConn{
		in:   chanServerToClient,
		out:  chanClientToServer,
		term: term,
	}

	clientDummyConn := &dummyConn{
		in:   chanClientToServer,
		out:  chanServerToClient,
		term: term,
	}

	serverDummyConn.partner = clientDummyConn
	clientDummyConn.partner = serverDummyConn

	return serverDummyConn, clientDummyConn
}

func (c *dummyConn) Read(p []byte) (n int, err error) {
	select {
	case <-c.term:
		// terminated
		return 0, io.EOF

	case b, ok := <-c.out:
		if !ok {
			return 0, io.EOF
		}
		n = copy(p, b)
		return
	}
}

func (c *dummyConn) Write(p []byte) (n int, err error) {
	select {
	case <-c.term:
		// terminated
		return 0, io.EOF
	case c.in <- p:
		return len(p), nil
	}
}

func (c *dummyConn) Close() error {
	select {
	case <-c.term:
		// already closed, do nothing
	default:
		// close location connections
		close(c.term)
		close(c.in)

		// signal partner to close
		c.partner.Close()
	}
	return nil
}

// NewDummySessions create a pair of server and client sessions
// that connects to each other with the given session ID and buffer size.
//
// Buffer size specify the the buffer of []byte arrays that is read / write
// to and from each other.
func NewDummySessions(sessID string, size int) (serverSess, clientSess *comms.Session) {
	serverConn, clientConn := NewDummyConns(size)
	serverSess = comms.NewSession(sessID, serverConn)
	clientSess = comms.NewSession(sessID, clientConn)
	return
}

type dummyMessageHandler struct {
	messages []comms.Message
	wg       *sync.WaitGroup
}

func (mh *dummyMessageHandler) HandleMessage(ctx context.Context, m comms.Message, mw comms.MessageWriter) error {
	mh.messages = append(mh.messages, m)
	mh.wg.Done()
	// do not emit any message.
	return nil
}

func (mh *dummyMessageHandler) Wait() {
	mh.wg.Wait()
}

func newDummyMessageHandler(wait int) *dummyMessageHandler {
	wg := &sync.WaitGroup{}
	wg.Add(wait)
	return &dummyMessageHandler{
		messages: make([]comms.Message, 0),
		wg:       wg,
	}
}

func TestNewDummySessions(t *testing.T) {

	// Make sure the dummy server and client session are created correctly
	// and they can communicate with each other.
	serverSess, clientSess := NewDummySessions("session-1", 1024)
	if serverSess == nil {
		t.Fatalf("server session is nil")
	}
	if clientSess == nil {
		t.Fatalf("client session is nil")
	}

	if serverSess.ID() != "session-1" {
		t.Errorf("server session ID is not correct. expected %#v, got %#v", "session-1", serverSess.ID())
	}
	if clientSess.ID() != "session-1" {
		t.Errorf("client session ID is not correct. expected %#v, got %#v", "session-1", clientSess.ID())
	}

	err := serverSess.WriteMessage(comms.NewSimpleMessage("session-1", "test:server-to-client"))
	if err != nil {
		t.Fatalf("unexpected error writing message: %s", err)
	}
	m, err := clientSess.ReadMessage()
	if err != nil {
		t.Fatalf("unexpected error reading message: %s", err)
	}
	if m.SessionID() != "session-1" {
		t.Errorf("message session ID is not correct. expected %#v, got %#v", "session-1", m.SessionID())
	}
	if m.Type() != "test:server-to-client" {
		t.Errorf("message type is not correct. expected %#v, got %#v", "test", m.Type())
	}

	err = clientSess.WriteMessage(comms.NewSimpleMessage("session-1", "test:client-to-server"))
	if err != nil {
		t.Fatalf("unexpected error writing message: %s", err)
	}
	m, err = serverSess.ReadMessage()
	if err != nil {
		t.Fatalf("unexpected error reading message: %s", err)
	}
	if m.SessionID() != "session-1" {
		t.Errorf("message session ID is not correct. expected %#v, got %#v", "session-1", m.SessionID())
	}
	if m.Type() != "test:client-to-server" {
		t.Errorf("message type is not correct. expected %#v, got %#v", "test", m.Type())
	}
}

func TestSimpleMessageQueue(t *testing.T) {
	sc := comms.NewSessionCollection()
	smq := comms.NewSimpleMessageQueue(sc, 0)

	// Dummy session for test
	s1, c1 := NewDummySessions("session-1", 0)
	defer s1.Close()
	defer c1.Close()
	s2, c2 := NewDummySessions("session-2", 0)
	defer s2.Close()
	defer c2.Close()

	// Add session after the queue is started
	mh := newDummyMessageHandler(2)
	smq.Start(mh, nil) // dummy message handle won't be using the message writer.
	defer smq.Stop()
	sc.Add(s1)
	sc.Add(s2)

	// Send a message from client c1
	err := c1.WriteMessage(comms.NewSimpleMessage("session-1", "test:client-to-server:1"))
	if err != nil {
		t.Fatalf("unexpected error writing message: %s", err)
	}

	// Send a message from client c2
	err = c2.WriteMessage(comms.NewSimpleMessage("session-2", "test:client-to-server:2"))
	if err != nil {
		t.Fatalf("unexpected error writing message: %s", err)
	}

	// Wait for the message handler to receive both messages
	mh.Wait()

	// Check messages received
	if want, have := 2, len(mh.messages); want != have {
		t.Fatalf("unexpected number of messages. want %d, have %d", want, have)
	}
	if want, have := "test:client-to-server:1", mh.messages[0].Type(); want != have {
		t.Errorf("unexpected message type. want %#v, have %#v", want, have)
	}
	if want, have := "test:client-to-server:2", mh.messages[1].Type(); want != have {
		t.Errorf("unexpected message type. want %#v, have %#v", want, have)
	}
}

func TestSimpleMessageBroker(t *testing.T) {

	// Create a new session collection
	sc := comms.NewSessionCollection()

	s1, c1 := NewDummySessions("session-1", 1024)
	s2, c2 := NewDummySessions("session-2", 1024)

	// Add the sessions to the collection
	sc.Add(s1)
	sc.Add(s2)

	// Create a new router
	r := comms.NewSimpleMessageBroker(sc)

	// Create a new response message
	err := r.WriteMessage(comms.NewSimpleMessage("session-1", "response"))
	if err != nil {
		t.Fatalf("unexpected error writing message: %s", err)
	}

	// Read the message from the client
	m, err := c1.ReadMessage()
	if err != nil {
		t.Fatalf("unexpected error reading message: %s", err)
	}
	if expected, actual := "session-1", m.SessionID(); expected != actual {
		t.Errorf("message session ID is not correct. expected %#v, got %#v", expected, actual)
	}
	if expected, actual := "response", m.Type(); expected != actual {
		t.Errorf("message type is not correct. expected %#v, got %#v", expected, actual)
	}

	// Create a new event message
	err = r.WriteMessage(comms.NewSimpleMessage("session-1", "event"))
	if err != nil {
		t.Fatalf("unexpected error writing message: %s", err)
	}

	// Read the message from the client c1
	m, err = c1.ReadMessage()
	if err != nil {
		t.Fatalf("unexpected error reading message: %s", err)
	}
	if expected, actual := "event", m.Type(); expected != actual {
		t.Errorf("message type is not correct. expected %#v, got %#v", expected, actual)
	}

	// Read the message from client c2
	m, err = c2.ReadMessage()
	if err != nil {
		t.Fatalf("unexpected error reading message: %s", err)
	}
	if expected, actual := "event", m.Type(); expected != actual {
		t.Errorf("message type is not correct. expected %#v, got %#v", expected, actual)
	}

	// Note: events are not session specific.
	//       it should either have a session ID of "" or nil.
}
