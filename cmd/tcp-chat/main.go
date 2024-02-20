package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	nc "nc/pkg/logger"
)

type Message struct {
	ID   string
	Name string
	Text string
}

var (
	clients    = make(map[string]net.Conn)
	messages   = make(chan Message)
	messageLog = []string{}
	logMutex   = sync.Mutex{}
)

var colors = map[int]string{
	0: "\033[0m",
	1: "\033[31m",
	2: "\033[32m",
	3: "\033[33m",
	4: "\033[34m",
	5: "\033[35m",
	6: "\033[36m",
}

func broadcaster() {
	for msg := range messages {
		logMutex.Lock()
		messageLog = append(messageLog, msg.Text)
		logMutex.Unlock()

		for clientName, conn := range clients {
			if clientName != msg.ID {
				fmt.Fprint(conn, msg.Text)
			}
		}
	}
}

func handleConnection(conn net.Conn, lg *nc.Logger) {
	welcomeMessage, err := os.ReadFile("hello.txt")
	if err != nil {
		lg.ErrorLog.Println("Error reading welcome message file:", err)
	}
	conn.Write(welcomeMessage)

	conn.Write([]byte("[ENTER YOUR NAME]: "))
	login, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		lg.ErrorLog.Println("Error reading client name:", err)
		conn.Close()
		return
	}

	login = strings.TrimSpace(login)
	if !isValidNickname(login) {
		conn.Write([]byte("Your name must contain at least 2, and less than 15 characters\nThe name must contain only letters and digits\n"))
		conn.Close()
		return
	}

	logMutex.Lock()
	if _, exists := clients[login]; exists {
		conn.Write([]byte("This name is already taken. Please reconnect with a different name.\n"))
		conn.Close()
		logMutex.Unlock()
		return
	}

	name := ColorfulNickname(login)
	clients[login] = conn
	logMutex.Unlock()

	messages <- Message{ID: login, Text: fmt.Sprintf("\n%s has joined our chat", name)}

	logMutex.Lock()

	for i, msg := range messageLog {

		if i == 1 {
			conn.Write([]byte(strings.TrimSpace(msg)))
			continue
		}

		if i == len(messageLog)-1 {
			conn.Write([]byte(msg + "\n"))
		}

	}

	logMutex.Unlock()

	scanner := bufio.NewScanner(conn)
	conn.Write([]byte(fmt.Sprintf("[%s][%s]:(1) ", time.Now().Format("2006-01-02 15:04:05"), name)))

	for scanner.Scan() {
		conn.Write([]byte(fmt.Sprintf("[%s][%s]: ", time.Now().Format("2006-01-02 15:04:05"), name)))

		text := scanner.Text()

		if isANSIMessage(text) {
			conn.Write([]byte("\nThis message contain ANSI symbols :("))
			continue
		}

		if text != "" {
			messages <- Message{ID: login, Text: fmt.Sprintf("\n[%s][%s]: %s", time.Now().Format("2006-01-02 15:04:05"), name, text)}
		}

	}

	if err := scanner.Err(); err != nil {
		lg.ErrorLog.Println("Error reading from client:", err)
	}

	logMutex.Lock()
	delete(clients, name)
	logMutex.Unlock()

	messages <- Message{Text: fmt.Sprintf("\n%s has left the chat", name)}
	lg.InfoLog.Printf("Client disconnected: %s\n", conn.RemoteAddr().String())
	conn.Close()
}

func isValidNickname(name string) bool {
	re := regexp.MustCompile(`^\w{2,15}$`)
	return re.MatchString(name)
}

func isANSIMessage(msg string) bool {
	for _, r := range msg {
		if r == '\x1B' {
			return true
		}
	}
	return false
}

func ColorfulNickname(name string) string {
	diceRoll := rand.Intn(len(colors)-1) + 1
	colorCode, exists := colors[diceRoll]
	if !exists {
		return name
	}
	return colorCode + name + colors[0]
}

func main() {
	lg := nc.NewLogger()

	port := "8989"
	if len(os.Args) == 2 {
		port = os.Args[1]
	}
	if len(os.Args) > 2 {
		fmt.Println("[USAGE]: ./tcp-chat $port")
		return
	}

	lg.InfoLog.Printf("Listening on the port :%s\n", port)
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		lg.ErrorLog.Fatalf("Error starting server: %v", err)
	}
	defer listener.Close()

	go broadcaster()

	for {
		conn, err := listener.Accept()
		if err != nil {
			lg.ErrorLog.Println("Error accepting connection:", err)
			continue
		}

		if len(clients) > 10 {
			conn.Write([]byte("limits of users is reached, try again later"))
			conn.Close()
		}

		lg.InfoLog.Printf("Client connected: %s\n", conn.RemoteAddr().String())

		go handleConnection(conn, lg)
	}
}
