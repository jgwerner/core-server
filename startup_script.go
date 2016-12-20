package core

import (
	"bytes"
	"database/sql"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

const scriptPath = "/start.sh"

// StartScript runs server startup script
func StartScript(dbURL string, serverID int) error {
	err := checkScriptInDB(dbURL, serverID)
	if err != nil {
		return err
	}
	script, err := getScriptData(scriptPath)
	if err != nil {
		return err
	}
	script = normalizeScriptData(script)
	err = ioutil.WriteFile(scriptPath, script, 0777)
	if err != nil {
		log.Printf("Script write error: %s", err)
		return err
	}
	cmd := exec.Command("bash", "-e", scriptPath)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		log.Printf("Script run error: %s", err)
		return err
	}
	return nil
}

func checkScriptInDB(dbURL string, serverID int) error {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("Error connecting database: %s", err)
		return err
	}
	defer db.Close()
	var dbscript string
	err = db.QueryRow(`SELECT startup_script FROM servers WHERE id = $1`, serverID).Scan(&dbscript)
	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		log.Println(err)
		return err
	}
	return nil
}

func getScriptData(scriptPath string) ([]byte, error) {
	script, err := ioutil.ReadFile(scriptPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []byte{}, nil
		}
		log.Printf("Script read error: %s", err)
		return []byte{}, err
	}
	return script, nil
}

func normalizeScriptData(data []byte) []byte {
	replaceWin := bytes.Replace(data, []byte("\r\n"), []byte("\n"), -1)
	return bytes.Replace(replaceWin, []byte("\r"), []byte("\n"), -1)
}
