package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestStartScript(t *testing.T) {
	content := []byte(`echo "test"`)
	tmpfile, err := ioutil.TempFile("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	_, err = tmpfile.Write(content)
	if err != nil {
		t.Fatal(err)
	}
	err = tmpfile.Close()
	if err != nil {
		t.Fatal(err)
	}
	scriptPath = tmpfile.Name()
	var buf bytes.Buffer
	out = &buf
	StartScript()
	out = os.Stderr
	if !bytes.Contains(buf.Bytes(), []byte("test")) {
		t.Error("wrong output")
	}
}

func TestGetScriptData(t *testing.T) {
	content := []byte(`content of the test file`)
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	_, err = tmpfile.Write(content)
	if err != nil {
		t.Fatal(err)
	}
	err = tmpfile.Close()
	if err != nil {
		t.Fatal(err)
	}
	cont, err := getScriptData(tmpfile.Name())
	if err != nil {
		t.Error(err)
	}
	if bytes.Compare(content, cont) != 0 {
		t.Errorf("File content does not match expected:\nExpected: %s\nActual: %s\n", content, cont)
	}
}
