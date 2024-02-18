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

// Определение структуры для сообщения
type Message struct {
	Name string // Имя отправителя
	Text string // Текст сообщения
}

var (
	clients    = make(map[string]net.Conn)
	messages   = make(chan Message) // Обновлен канал для передачи структуры Message
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
			if clientName != msg.Name { // Исключить отправителя
				fmt.Fprintln(conn, msg.Text)
			}
		}
	}
}

func handleConnection(conn net.Conn, lg *nc.Logger) {
	wlcmsg, err := os.ReadFile("hello.txt")
	if err != nil {
		lg.ErrorLog.Println("Error reading welcome message file:", err)
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
	if !isValidNickname(name) {
		conn.Write([]byte("Your name must contain at least 2, and less than 15 characters\n"))
		conn.Write([]byte("The name must contain only letters and digits\n"))
		conn.Close()
		return
	}

	logMutex.Lock()
	if _, exists := clients[name]; exists {
		conn.Write([]byte("This name is already taken. Please reconnect with a different name.\n"))
		conn.Close()
		logMutex.Unlock()
		return
	}

	name = ColorfulNickname(name)
	clients[name] = conn
	logMutex.Unlock()

	messages <- Message{Name: name, Text: fmt.Sprintf("\n%s has joined our chat", name)}

	logMutex.Lock()
	for _, msg := range messageLog {
		conn.Write([]byte(msg + "\n"))
	}
	logMutex.Unlock()

	scanner := bufio.NewScanner(conn)
	conn.Write([]byte(fmt.Sprintf("[%s][%s]:(1) ", time.Now().Format("2006-01-02 15:04:05"), name)))

	for scanner.Scan() {
		conn.Write([]byte(fmt.Sprintf("[%s][%s]: ", time.Now().Format("2006-01-02 15:04:05"), name)))

		text := scanner.Text()
		if text != "" {
			messages <- Message{Name: name, Text: fmt.Sprintf("[%s][%s]: %s", time.Now().Format("2006-01-02 15:04:05"), name, text)}
		}

		if text == "" {
			continue
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
		lg.InfoLog.Printf("Client connected: %s\n", conn.RemoteAddr().String())

		go handleConnection(conn, lg)
	}
}
