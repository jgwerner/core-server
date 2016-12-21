package core

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"golang.org/x/net/websocket"
)

func TestMain(m *testing.M) {
	RunKernelGateway(os.Stdout, os.Stderr, "python")
	time.Sleep(1 * time.Second)
	GetKernel()
	exitCode := m.Run()
	shutdownCurrentKernel()
	syscall.Kill(kgPID, syscall.SIGTERM)
	os.Exit(exitCode)
}

func TestGetKernel(t *testing.T) {
	if currentKernel.ID == "" {
		t.Error("Wrong kernel id")
	}
}

func TestRun_Success(t *testing.T) {
	expected := `'{"test": 1}'`
	code := fmt.Sprintf(`def test():
	return %s
test()`, expected)
	respChan, errChan := Run(code)
	var data string
	select {
	case data = <-respChan:
		if data != expected {
			t.Errorf("Wrong data\nExpected: %s\nActual: %s\n", expected, data)
		}
	case data = <-errChan:
		t.Error(data)
	}
}

func TestRun_Stream(t *testing.T) {
	assert := func(r io.Reader, msg string) {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		if !strings.Contains(buf.String(), msg) {
			t.Error("Wrong msg")
		}
	}

	const stdoutMsg = "streamstdouttest"
	const stderrMsg = "streamstderrtest"

	code := fmt.Sprintf(`import sys
def test():
	sys.stdout.write('%s')
	sys.stderr.write('%s')
	return '{}'
test()`, stdoutMsg, stderrMsg)
	go assert(os.Stdout, stdoutMsg)
	go assert(os.Stdout, stderrMsg)
	respChan, errChan := Run(code)
	select {
	case <-respChan:
	case <-errChan:
	}

}

func TestRun_Fail(t *testing.T) {
	code := `test1`
	respChan, errChan := Run(code)
	var data string
	select {
	case data = <-respChan:
		t.Error("No error with bad data")
	case data = <-errChan:
		if data == "" {
			t.Error("No traceback")
		}
	}
}

func shutdownCurrentKernel() {
	ws := dialKernelWebSocket()
	defer ws.Close()
	shutdownMsg := createMsg("shutdown_request", "shell", map[string]interface{}{
		"restart": false,
	})
	err := websocket.JSON.Send(ws, &shutdownMsg)
	if err != nil {
		log.Println("Kernel shutdown error", err)
	}
	for {
		var reply msg
		err = websocket.JSON.Receive(ws, &reply)
		if err != nil {
			log.Println("Kernel shutdown error", err)
		}
		switch reply.Header.MsgType {
		case "shutdown_reply":
			return
		}
	}
}
