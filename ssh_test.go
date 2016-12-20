package core

import (
	"testing"
)

func TestEndpoint_String(t *testing.T) {
	e := endpoint{host: "localhost", port: 1001}
	if e.String() != "localhost:1001" {
		t.Error("Wrong endpoint string")
	}
}
