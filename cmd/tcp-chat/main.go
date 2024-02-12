package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	nc "nc/pkg/logger"
)

var (
	clients    = make(map[net.Conn]*Client)
	messages   = make(chan string)
	messageLog = []string{}
	logMutex   = sync.Mutex{}
)

type Client struct {
	name string
	ch   chan string
}

func broadcaster(lg *nc.Logger) {
	for msg := range messages {
		logMutex.Lock()
		messageLog = append(messageLog, msg)
		lg.ChatLog.Println(msg)
		logMutex.Unlock()

		for _, cli := range clients {
			cli.ch <- msg
		}
	}
}

func handleConnection(conn net.Conn, lg *nc.Logger) {
	ch := make(chan string)
	go clientWriter(conn, ch)

	wlcmsg, err := os.ReadFile("hello.txt")
	if err != nil {
		lg.ErrorLog.Println("Error reading file:", err)
	}
	conn.Write([]byte(wlcmsg))

	conn.Write([]byte("[ENTER YOUR NAME]: "))
	name, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		lg.ErrorLog.Println("Error reading client name:", err)
		conn.Close()
		return
	}
	name = strings.TrimSpace(name)

	clients[conn] = &Client{name: name, ch: ch}
	messages <- fmt.Sprintf("%s has joined our chat", name)

	logMutex.Lock()
	for _, msg := range messageLog {
		ch <- msg
	}
	logMutex.Unlock()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		messages <- fmt.Sprintf("[%s][%s]: %s", time.Now().Format("2006-01-02 15:04:05"), name, text)
	}

	delete(clients, conn)
	messages <- fmt.Sprintf("%s has left the chat", name)
	lg.InfoLog.Printf("Client disconnected: %s\n", conn.RemoteAddr().String())
	conn.Close()
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}

func main() {
	lg := nc.NewLogger()

	port := "8989"
	if len(os.Args) == 2 {
		port = os.Args[1]
	}

	if len(os.Args) > 2 {
		fmt.Println("[USAGE]: ./cmd/tcp-chat/ $port")
		return
	}

	lg.InfoLog.Printf("Listening on the port :%s\n", port)
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		lg.ErrorLog.Fatalf("Error starting server: %v", err)
	}
	defer listener.Close()

	go broadcaster(lg)

	for len(clients) < 20 {
		conn, err := listener.Accept()
		if err != nil {
			lg.ErrorLog.Println("Error accepting connection:", err)
			continue
		}

		clientAddr := conn.RemoteAddr().String()
		lg.InfoLog.Printf("Client connected: %s\n", clientAddr)

		go handleConnection(conn, lg)
	}
}
