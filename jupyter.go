package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/satori/go.uuid"
	"golang.org/x/net/websocket"
)

// domain is default jupyter kernel gateway listening address
// TODO: port should not be hardcoded
const domain = "localhost:8888"

var (
	baseURI       = fmt.Sprintf("http://%s", domain)
	wsURI         = fmt.Sprintf("ws://%s", domain)
	currentKernel = kernel{Name: "python"}
	kgPID         int
)

// msg is jupyter message implementation
// https://jupyter-client.readthedocs.io/en/latest/messaging.html#general-message-format
type msg struct {
	Header       *header                `json:"header"`
	ParentHeader *header                `json:"parent_header"`
	Channel      string                 `json:"channel"`
	Content      map[string]interface{} `json:"content"`
	Metadata     map[string]interface{} `json:"metadata"`
	Buffers      []interface{}          `json:"buffers"`
}

type header struct {
	Username string    `json:"username"`
	Version  string    `json:"version"`
	Session  uuid.UUID `json:"session"`
	MsgID    uuid.UUID `json:"msg_id"`
	MsgType  string    `json:"msg_type"`
	Date     string    `json:"date"`
}

// kernel represents jupyter kernel info
type kernel struct {
	Name string `json:"name"`
	ID   string `json:"id,omitempty"`
}

// RunKernelGateway runs jupyter kernel gateway
// https://github.com/jupyter/kernel_gateway
func RunKernelGateway(stdout, stderr io.Writer, kernelName string) {
	currentKernel.Name = kernelName
	cmd := exec.Command(
		"jupyter",
		"kernelgateway",
	)
	cmd.Stderr = stdout
	cmd.Stdout = stderr
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	kgPID = cmd.Process.Pid
	go func() {
		log.Printf("Error starting kernel gateway: %s", cmd.Wait())
	}()
}

// Run sends code to jupyter kernel for processing
func Run(ctx context.Context, stats *Stats, code string) (string, error) {
	ws := dialKernelWebSocket()
	err := websocket.JSON.Send(ws, createExecuteMsg(code))
	if err != nil {
		return "", err
	}
	respCh := make(chan string)
	errCh := make(chan string)
	go handleWebSocket(ws, respCh, errCh)
	var data string
	stats.Start = time.Now().UTC()
	select {
	case data = <-respCh:
		stats.ExitCode = 0
		data = strings.Trim(data, "'")
	case data = <-errCh:
		stats.ExitCode = 1
		err = errors.New("Script error")
	case <-ctx.Done():
	}
	stats.Stop = time.Now().UTC()
	return data, err
}

// SetKernelName sets currentKernel name
func SetKernelName(name string) {
	currentKernel.Name = name
}

// GetKernel gets kernel id by name and starts kernel process
func GetKernel() {
	var body bytes.Buffer
	json.NewEncoder(&body).Encode(&currentKernel)
	credentials := url.Values{}
	credentials.Set("auth_username", "fakeuser")
	credentials.Set("auth_password", "fakepass")
	uri, _ := url.Parse(fmt.Sprintf(`%s/api/kernels`, baseURI))
	uri.RawQuery = credentials.Encode()
	// TODO: this could be handled better
	time.Sleep(2 * time.Second)
	response, err := http.Post(uri.String(), "application/json", &body)
	if err != nil {
		log.Println(err)
		return
	}
	if response != nil {
		defer response.Body.Close()
	}
	err = json.NewDecoder(response.Body).Decode(&currentKernel)
	if err != nil {
		log.Printf("Error decoding kernel: %s", err)
	}
}

// createMsg creates msg to be sent to kernel gateway
func createMsg(msgType, channel string, content map[string]interface{}) *msg {
	return &msg{
		Header: &header{
			Version: "5.0",
			MsgID:   uuid.NewV4(),
			MsgType: msgType,
			Session: uuid.NewV4(),
			Date:    time.Now().Format(time.RFC3339),
		},
		Channel: channel,
		Content: content,
	}
}

func createExecuteMsg(code string) *msg {
	return createMsg("execute_request", "shell", map[string]interface{}{
		"code":          code,
		"silent":        false,
		"store_history": false,
		"allow_stdin":   false,
	})
}

// handleWebSocket handles jupyter gateway websocket connection
func handleWebSocket(ws *websocket.Conn, respCh chan string, errCh chan string) {
	var err error
	defer ws.Close()
	for {
		var respMsg msg
		err = websocket.JSON.Receive(ws, &respMsg)
		if err != nil {
			if err == io.EOF {
				ws.Close()
				break
			}
			log.Printf("Error receiving message from websocket: %s", err)
		} else {
			go handleResponseMsg(&respMsg, respCh, errCh)
		}
	}
}

// handleResponseMsg handles required types of jupyter messages
func handleResponseMsg(respMsg *msg, resp chan string, errCh chan string) {
	var err error
	switch respMsg.Header.MsgType {
	case "display_data", "execute_result":
		data := respMsg.Content["data"].(map[string]interface{})
		for _, v := range data {
			resp <- v.(string)
			break
		}
		break
	case "stream":
		var out io.Writer
		outMsg := respMsg.Content["text"].(string)
		switch respMsg.Content["name"].(string) {
		case "stdout":
			out = os.Stdout
		case "stderr":
			out = os.Stderr
		}
		_, err = fmt.Fprint(out, outMsg)
		if err != nil {
			log.Println(err)
		}
	case "error":
		var buf bytes.Buffer
		for _, v := range respMsg.Content["traceback"].([]interface{}) {
			buf.WriteString(v.(string))
		}
		errCh <- buf.String()
		break
	case "execute_reply":
		resp <- ""
		break
	}
}

// dialKernelWebSocket is a helper function to quick message sending
func dialKernelWebSocket() *websocket.Conn {
	if currentKernel.ID == "" {
		GetKernel()
	}
	uri := fmt.Sprintf("%s/api/kernels/%s/channels", wsURI, currentKernel.ID)
	ws, err := websocket.Dial(uri, "", baseURI)
	if err != nil {
		log.Fatalf("Error dialing websocket: %s", err)
	}
	return ws
}
