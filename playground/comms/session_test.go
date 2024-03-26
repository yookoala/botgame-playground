package comms_test

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net"
	"sync"
	"testing"

	"github.com/yookoala/botgame-playground/playground/comms"
)

func startDummyServer(t *testing.T, network, address string) net.Listener {
	l, err := net.Listen(network, address)
	if err != nil {
		log.Fatalf("unexpected error creating socket to listen: %#v", err)
	}

	go func(l net.Listener) {
		for {
			conn, err := l.Accept()
			if err != nil {
				if _, ok := err.(*net.OpError); ok {
					// socket closed. simply quit.
					return
				}
				log.Fatalf("unexpected error listening to incomming connection: %s", err)
			}

			go func(conn net.Conn) {
				r := bufio.NewReader(conn)
				b, _ := r.ReadBytes('\n')
				t.Logf("dummy server received: %s", b)
			}(conn)

			go func(conn net.Conn) {
				w := json.NewEncoder(conn)
				m := comms.NewGreeting("1")
				w.Encode(m)
				t.Logf("dummy server sent: %#v", m)
			}(conn)
		}
	}(l)
	return l
}

// Test if the session implementation can be read and write simultaneously.
func TestSessionSymultanousIO(t *testing.T) {
	start := &sync.WaitGroup{}
	start.Add(1)

	end := &sync.WaitGroup{}
	end.Add(2)

	socketName := "./session_test.sock"

	l := startDummyServer(t, "unix", socketName)
	defer l.Close()

	conn, err := net.Dial("unix", socketName)
	if err != nil {
		t.Fatalf("unexpected error connecting to server: %s", err)
	}

	_ = conn

	sess := comms.NewSession("1", conn)

	go func(sess *comms.Session) {
		start.Wait()
		m, err := sess.ReadMessage()
		if err != nil {
			t.Logf("client experience unexpected error reading message: %#v", err)
			t.Fail()
		}
		t.Logf("client session received: %s", m)
		end.Done()
	}(sess)

	go func(sess *comms.Session) {
		start.Wait()
		m := comms.MustMessage(comms.NewMessageFromJSONString(`{"session_id":"1","type":"message","content":"hello"}`))
		err := sess.WriteMessage(m)
		if err != nil {
			t.Logf("client experience unexpected error writing message: %#v", err)
			t.Fail()
		}
		t.Logf("client session sent: %s", m)
		end.Done()
	}(sess)

	start.Done()
	end.Wait()
}

func TestSessionCollection(t *testing.T) {
	sc := comms.NewSessionCollection()
	if sc == nil {
		t.Fatalf("unexpected nil session collection")
	}

	var dummyConn io.ReadWriteCloser
	var count int

	// Count the number of session added, if any.
	onAdd := func(s *comms.Session) {
		count++
	}
	sc.OnAdd(onAdd)

	// New dummy session for test
	s1 := comms.NewSession("1", dummyConn)

	// Test if session collection is empty
	if expected, actual := false, sc.Has("1"); expected != actual {
		t.Errorf("expected %#v, got %#v", expected, actual)
	}
	if expected, actual := (*comms.Session)(nil), sc.Get("1"); expected != actual {
		t.Errorf("expected %#v, got %#v", expected, actual)
	}
	if expected, actual := 0, sc.Len(); expected != actual {
		t.Errorf("expected %#v, got %#v", expected, actual)
	}

	// Test adding a session to the collection.
	if err := sc.Add(s1); err != nil {
		t.Errorf("unexpected error adding session to collection: %#v", err)
	}
	if expected, actual := true, sc.Has("1"); expected != actual {
		t.Errorf("expected %#v, got %#v", expected, actual)
	}
	if expected, actual := s1, sc.Get("1"); expected != actual {
		t.Errorf("expected %#v, got %#v", expected, actual)
	}
	if expected, actual := 1, sc.Len(); expected != actual {
		t.Errorf("expected %#v, got %#v", expected, actual)
	}
	if expected, actual := 1, count; expected != actual {
		t.Errorf("expected %#v, got %#v", expected, actual)
	}

	// Test to add another session with duplicated id
	s1b := comms.NewSession("1", dummyConn)
	if err := sc.Add(s1b); err == nil {
		t.Errorf("expected error adding session with duplicated id")
	}

	// Test removing session from collection.
	sc.Remove("1")
	if expected, actual := false, sc.Has("1"); expected != actual {
		t.Errorf("expected %#v, got %#v", expected, actual)
	}
	if expected, actual := (*comms.Session)(nil), sc.Get("1"); expected != actual {
		t.Errorf("expected %#v, got %#v", expected, actual)
	}
	if expected, actual := 0, sc.Len(); expected != actual {
		t.Errorf("expected %#v, got %#v", expected, actual)
	}
}
