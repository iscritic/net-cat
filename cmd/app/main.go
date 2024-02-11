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
	clients    = make(map[net.Conn]*Client) // Все подключенные клиенты
	messages   = make(chan string)          // Все сообщения для рассылки, включая уведомления о подключении/отключении
	messageLog = []string{}                 // История сообщений
	logMutex   = sync.Mutex{}               // Mutex для безопасного доступа к истории сообщений
)

type Client struct {
	name string
	ch   chan string // канал отправки сообщений клиенту
}

func broadcaster() {
	for msg := range messages {
		logMutex.Lock()
		messageLog = append(messageLog, msg) // Сохраняем сообщение в истории
		logMutex.Unlock()

		for _, cli := range clients {
			cli.ch <- msg
		}
	}
}

func handleConnection(conn net.Conn, lg *nc.Logger) {
	ch := make(chan string) // канал исходящих сообщений для клиента
	go clientWriter(conn, ch)

	// Запрос имени и регистрация клиента
	conn.Write([]byte("[ENTER YOUR NAME]: "))
	name, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		lg.ErrorLog.Println("Error reading client name:", err)
		conn.Close()
		return
	}
	name = strings.TrimSpace(name)

	clients[conn] = &Client{name: name, ch: ch}
	messages <- fmt.Sprintf("%s has joined our chat", name) // Уведомление о новом пользователе

	// Отправка истории сообщений новому клиенту
	logMutex.Lock()
	for _, msg := range messageLog {
		ch <- msg
	}
	logMutex.Unlock()

	// Чтение и отправка сообщений от клиента
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" { // Пропускаем пустые сообщения
			continue
		}
		messages <- fmt.Sprintf("[%s][%s]: %s", time.Now().Format("2006-01-02 15:04:05"), name, text)
	}

	// Уведомление об отключении пользователя и его удаление из списка
	delete(clients, conn)
	messages <- fmt.Sprintf("%s has left the chat", name)
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

		go handleConnection(conn, lg)
	}
}
