package integration

import "testing"

func TestServerUp(t *testing.T) {
	stopCh, _, _, err := StartDefaultServer()
	if err != nil {
		t.Fatal(err)
	}
	defer close(stopCh)
}
