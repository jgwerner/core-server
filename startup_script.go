package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
)

var scriptPath = "/start.sh"

// StartScript runs server startup script
func StartScript() error {
	script, err := getScriptData(scriptPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	script = normalizeScriptData(script)
	err = ioutil.WriteFile(scriptPath, script, 0777)
	if err != nil {
		return err
	}
	cmd := exec.Command("bash", "-e", scriptPath)
	cmd.Stderr = out
	cmd.Stdout = out
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func getScriptData(scriptPath string) ([]byte, error) {
	script, err := ioutil.ReadFile(scriptPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []byte{}, err
		}
		logger.Printf("Script read error: %s", err)
		return []byte{}, err
	}
	return script, nil
}

func normalizeScriptData(data []byte) []byte {
	replaceWin := bytes.Replace(data, []byte("\r\n"), []byte("\n"), -1)
	return bytes.Replace(replaceWin, []byte("\r"), []byte("\n"), -1)
}
