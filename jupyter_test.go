package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"golang.org/x/net/websocket"
)

func TestMain(m *testing.M) {
	out = new(bytes.Buffer)
	RunKernelGateway(out, out, "python")
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

func prepareScript(code string) error {
	args.ResourceDir = "/tmp"
	args.Script = "test.py"
	scriptPath := filepath.Join(args.ResourceDir, args.Script)
	return ioutil.WriteFile(scriptPath, []byte(code), 0644)
}

//func TestRun_Success(t *testing.T) {
//	expected := `{'test': 1}`
//	code := fmt.Sprintf("def test():\n\treturn %s", expected)
//	prepareScript(code)
//	data, _, _ := Run(context.Background(), args.Script, `test()`)
//	if data != expected {
//		t.Errorf("Wrong data\nExpected: %s\nActual: %s\n", expected, data)
//	}
//}

//func TestRun_Fail(t *testing.T) {
//	code := `test1`
//	prepareScript(code)
//	data, _, err := Run(context.Background(), args.Script, `test()`)
//	if err == nil {
//		t.Error("No error with bad data")
//	}
//	if data == "" {
//		t.Error("No traceback")
//	}
//}

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
