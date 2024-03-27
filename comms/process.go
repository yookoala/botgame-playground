package comms

import (
	"io"
	"os"
	"os/signal"
)

// CloseOnSignal closes the Closer when the specified OS
// signal is send to the process by the system.
func CloseOnSignal(cl io.Closer, sig ...os.Signal) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, sig...)
	go func() {
		<-ch
		cl.Close()
		close(ch)
	}()
}
