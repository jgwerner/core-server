package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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

func isJupyterRunning() bool {
	_, err := http.Get(fmt.Sprintf("%s/api", baseURI))
	if err != nil {
		return false
	}
	return true
}

// RunKernelGateway runs jupyter kernel gateway
// https://github.com/jupyter/kernel_gateway
func RunKernelGateway(stdout, stderr io.Writer, kernelName string) {
	if isJupyterRunning() {
		return
	}
	currentKernel.Name = kernelName
	cmd := exec.Command(
		"jupyter",
		"--NotebookApp.token=''",
		"kernelgateway",
		"--JupyterWebsocketPersonality.list_kernels=True",
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
func Run(ctx context.Context, stats *Stats, script, function string) (string, error) {
	ws := dialKernelWebSocket()
	respCh := make(chan string)
	errCh := make(chan string)
	go handleWebSocket(ws, respCh, errCh)
	scriptContent, err := scriptContent(script)
	if err != nil {
		return "", err
	}
	err = websocket.JSON.Send(ws, createExecuteMsg(scriptContent))
	if err != nil {
		return "", err
	}
	err = websocket.JSON.Send(ws, createExecuteMsg(function))
	if err != nil {
		return "", err
	}
	var data string
	stats.Start = time.Now().UTC()
	select {
	case data = <-respCh:
		stats.ExitCode = 0
		data = strings.Trim(data, "'")
	case data = <-errCh:
		stats.ExitCode = 1
		err = errors.New("Script error")
		break
	case <-ctx.Done():
	}
	stats.Stop = time.Now().UTC()
	return data, err
}

func scriptContent(script string) (string, error) {
	path := filepath.Join(args.ResourceDir, script)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SetKernelName sets currentKernel name
func SetKernelName(name string) {
	currentKernel.Name = name
}

func getKernelUri() string {
	return fmt.Sprintf(`%s/api/kernels`, baseURI)
}

func isKernelRunning(name string) bool {
	uri := getKernelUri()
	resp, err := http.Get(uri)
	if err != nil {
		return false
	}
	if resp != nil {
		defer resp.Body.Close()
	}
	runningKernels := []kernel{}
	err = json.NewDecoder(resp.Body).Decode(&runningKernels)
	if err != nil {
		return false
	}
	for _, kernel := range runningKernels {
		if kernel.Name == name {
			return true
		}
	}
	return false
}

func startKernel(k *kernel) {
	var body bytes.Buffer
	json.NewEncoder(&body).Encode(k)
	uri := getKernelUri()
	// TODO: this could be handled better
	time.Sleep(2 * time.Second)
	response, err := http.Post(uri, "application/json", &body)
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

// GetKernel gets kernel id by name and starts kernel process
func GetKernel() {
	if !isKernelRunning(currentKernel.Name) {
		startKernel(&currentKernel)
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
