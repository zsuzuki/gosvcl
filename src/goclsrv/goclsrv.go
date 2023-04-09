package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"os/exec"
	"os/user"
	"path/filepath"
)

var (
	execToken string
)

type Command struct {
	WorkingDir string `json:"working_dir"`
	CmdLine    string `json:"cmd_line"`
	Token      string `json:"token"`
}

func main() {
	port := flag.String("port", "33456", "The server port")
	token := flag.String("token", "-", "token")
	flag.Parse()

	usr, err := user.Current()
	if err != nil {
		fmt.Println("Error getting current user:", err)
		return
	}

	certPath := filepath.Join(usr.HomeDir, ".goclsrv/cert.pem")
	keyPath := filepath.Join(usr.HomeDir, ".goclsrv/key.pem")

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		fmt.Printf("Error loading certificate and key from %s and %s: %v\n", certPath, keyPath, err)
		return
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	execToken = *token

	listener, err := tls.Listen("tcp", ":"+*port, config)
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer listener.Close()

	fmt.Printf("Listening on :%s...\n", *port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	command := Command{}
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&command); err != nil {
		fmt.Println("Error decoding command:", err)
		return
	}
	if command.Token != execToken {
		fmt.Println("Error illegal token")
		return
	}

	cmd := exec.Command("sh", "-c", command.CmdLine)
	if command.WorkingDir != "" {
		cmd.Dir = command.WorkingDir
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("Error creating stdout pipe:", err)
		return
	}
	defer stdout.Close()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println("Error creating stderr pipe:", err)
		return
	}
	defer stderr.Close()

	if err := cmd.Start(); err != nil {
		fmt.Println("Error starting command:", err)
		return
	}

	go io.Copy(conn, stdout)
	go io.Copy(conn, stderr)

	if err := cmd.Wait(); err != nil {
		fmt.Println("Error waiting for command to finish:", err)
		return
	}
}
