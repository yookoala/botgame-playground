package comms_test

import (
	"bufio"
	"encoding/json"
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
