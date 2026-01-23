package signals

import (
	"testing"
	"time"
)

func TestSignalHandler(t *testing.T) {
	// init
	onlyOneSignalHandler = make(chan struct{})

	stopCh := SetupSignalHandler()
	RequestShutdown()

	ch := time.Tick(time.Second)
	select {
	case <-stopCh:

	case <-ch:
		t.Error("stopCh is not trigged")
	}
}

func TestSignalContext(t *testing.T) {
	// init
	onlyOneSignalHandler = make(chan struct{})

	ctx := SetupSignalContext()
	RequestShutdown()

	ch := time.Tick(time.Second)
	select {
	case <-ctx.Done():

	case <-ch:
		t.Error("stopCh is not trigged")
	}
}
