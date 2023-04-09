package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

type Command struct {
	WorkingDir string `json:"working_dir"`
	CmdLine    string `json:"cmd_line"`
	Token      string `json:"token"`
}

func main() {
	port := flag.String("port", "33456", "The server port")
	dir := flag.String("dir", "", "The working directory on the server for executing the command")
	token := flag.String("token", "-", "token")
	flag.Parse()

	if len(flag.Args()) < 2 {
		fmt.Println("Usage: client <server-name> -port <port> <command>")
		return
	}

	serverName := flag.Arg(0)
	cmd := strings.Join(flag.Args()[1:], " ")

	command := Command{
		WorkingDir: *dir,
		CmdLine:    cmd,
		Token:      *token,
	}

	commandBytes, err := json.Marshal(command)
	if err != nil {
		fmt.Println("Error marshaling command:", err)
		return
	}

	serverAddr := fmt.Sprintf("%s:%s", serverName, *port)

	config := &tls.Config{
		InsecureSkipVerify: true, // 本番環境ではCA認証を使用し、このオプションを削除してください
	}

	conn, err := tls.Dial("tcp", serverAddr, config)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	if _, err := conn.Write([]byte(commandBytes)); err != nil {
		fmt.Println("Error sending command:", err)
		return
	}

	if _, err := io.Copy(os.Stdout, conn); err != nil {
		fmt.Println("Error receiving output from server:", err)
		return
	}
}
