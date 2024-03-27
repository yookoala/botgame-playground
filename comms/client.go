package comms

import (
	"context"
	"io"
	"log"
	"net"
)

func StartClient(mh MessageHandler, conn net.Conn) (err error) {

	sessionID := ""
	tempSess := NewSession("", conn)
	sess := tempSess

	// Signal message handler to initialize.
	err = mh.HandleMessage(context.Background(), NewSignal("client:init", nil), tempSess)
	if err != nil {
		log.Fatal(err)
	}

	for {
		m, err := sess.ReadMessage()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Printf("unexpected read error: %s", err)
			continue
		}

		if sess == tempSess && m.SessionID() == "" {
			sessionID = m.SessionID()
			sess = NewSession(sessionID, conn)
		}

		err = mh.HandleMessage(WithSessionID(context.Background(), sessionID), m, sess)
		if err != nil {
			log.Printf("unexpected handle error: %s", err)
		}
	}
}
