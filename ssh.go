package core

import (
	"fmt"
	"io"
	"net"

	"log"

	"database/sql"
	"io/ioutil"
	"path"
	"strconv"
	"strings"

	cssh "golang.org/x/crypto/ssh"
)

// CreateSSHTunnels creates defined in db ssh tunnels
func CreateSSHTunnels(dbURL, resourceDir string, serverID int) error {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return err
	}
	defer db.Close()
	rows, err := db.Query(
		`SELECT
		  ssh_tunnel.local_port,
		  ssh_tunnel.host,
		  ssh_tunnel.remote_port,
		  ssh_tunnel.endpoint,
		  ssh_tunnel.username
		FROM ssh_tunnel
		WHERE ssh_tunnel.server_id = $1`, serverID)

	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var endpt string
		tunnel := sshTunnel{
			local:  &endpoint{},
			server: &endpoint{},
			remote: &endpoint{},
			config: &cssh.ClientConfig{},
		}
		err := rows.Scan(
			&tunnel.local.port,
			&tunnel.server.host,
			&tunnel.server.port,
			&endpt,
			&tunnel.config.User,
		)
		if err != nil {
			return err
		}
		tunnel.local.host = "0.0.0.0"
		sshKeyAuth, err := getSSHKeyAuthMethod(resourceDir)
		if err != nil {
			return err
		}
		tunnel.config.Auth = []cssh.AuthMethod{sshKeyAuth}
		splitEndpoint := strings.Split(endpt, ":")
		tunnel.remote.host = splitEndpoint[0]
		tunnel.remote.port, err = strconv.Atoi(splitEndpoint[1])
		if err != nil {
			return err
		}
		go tunnel.Start()
	}
	return rows.Err()
}

func getSSHKeyAuthMethod(resourceDir string) (cssh.AuthMethod, error) {
	key, err := ioutil.ReadFile(path.Join(resourceDir, ".ssh", "id_rsa"))
	if err != nil {
		return nil, err
	}
	signer, err := cssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}
	return cssh.PublicKeys(signer), nil
}

type endpoint struct {
	host string
	port int
}

func (e *endpoint) String() string {
	return fmt.Sprintf("%s:%d", e.host, e.port)
}

type sshTunnel struct {
	local  *endpoint
	server *endpoint
	remote *endpoint

	config *cssh.ClientConfig
}

func (tunnel *sshTunnel) Start() error {
	listener, err := net.Listen("tcp", tunnel.local.String())
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go tunnel.forward(conn)
	}
}

func (tunnel *sshTunnel) forward(localConn net.Conn) {
	serverConn, err := cssh.Dial("tcp", tunnel.server.String(), tunnel.config)
	handleError(err, "Server dial error")
	remoteConn, err := serverConn.Dial("tcp", tunnel.remote.String())
	handleError(err, "Remote dial error")

	copyConn := func(writer, reader net.Conn) {
		_, err := io.Copy(writer, reader)
		handleError(err, "io.Copy error")
	}

	go copyConn(localConn, remoteConn)
	go copyConn(remoteConn, localConn)
}

func handleError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s\n", msg, err)
	}
}
