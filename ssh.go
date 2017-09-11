package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"path"
	"strconv"
	"strings"

	"github.com/3Blades/go-sdk/client/projects"
	cssh "golang.org/x/crypto/ssh"
)

// CreateSSHTunnels creates defined in db ssh tunnels
func CreateSSHTunnels(args *Args) error {
	sshKeyAuth, err := getSSHKeyAuthMethod(args.ResourceDir)
	if err != nil {
		return err
	}
	cli := NewAPIClient(args.ApiRoot, args.ApiKey)
	params := projects.NewProjectsServersSSHTunnelsListParams()
	params.SetNamespace(args.Namespace)
	params.SetProject(args.ProjectID)
	params.SetServer(args.ServerID)
	res, err := cli.Projects.ProjectsServersSSHTunnelsList(params, cli.AuthInfo)
	if err != nil {
		return err
	}
	for _, apiTunnel := range res.Payload {
		splitEndpoint := strings.Split(*apiTunnel.Endpoint, ":")
		tunnel := &sshTunnel{
			local: &endpoint{
				port: int(*apiTunnel.LocalPort),
				host: "0.0.0.0",
			},
			server: &endpoint{
				host: *apiTunnel.Host,
				port: int(*apiTunnel.RemotePort),
			},
			remote: &endpoint{
				host: splitEndpoint[0],
			},
			config: &cssh.ClientConfig{
				User: *apiTunnel.Username,
				Auth: []cssh.AuthMethod{sshKeyAuth},
			},
		}
		tunnel.remote.port, err = strconv.Atoi(splitEndpoint[1])
		if err != nil {
			return err
		}
		go tunnel.start()
	}
	return nil
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

func (tunnel *sshTunnel) start() error {
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
		logger.Fatalf("%s: %s\n", msg, err)
	}
}
