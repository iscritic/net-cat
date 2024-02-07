package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	nc "nc/pkg/logger"
)

var messages = make(chan string)
var clients []*Client

type Client struct {
	conn net.Conn
	name string
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

	// Обработка входящих сообщений от клиентов
	go broadcastMessages()

	for {
		conn, err := listener.Accept()
		if err != nil {
			lg.ErrorLog.Println("Error accepting connection:", err)
			continue
		}

		clientAddr := conn.RemoteAddr().String()
		lg.InfoLog.Printf("Client connected: %s\n", clientAddr)

		// Спросить имя у клиента
		conn.Write([]byte("Enter your name: "))
		name, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			lg.ErrorLog.Println("Error reading client name:", err)
			continue
		}
		name = strings.TrimSpace(name)

		// Создать объект клиента и добавить его в список
		client := &Client{
			conn: conn,
			name: name,
		}
		clients = append(clients, client)

		// Приветствие нового клиента
		conn.Write([]byte("Welcome to the chat, " + name + "!\n"))

		// Обработка клиента в отдельной горутине
		go handleConnection(client)
	}
}

// Функция для обработки входящих сообщений от клиентов
func broadcastMessages() {
	for {
		// Получаем сообщение из канала
		message := <-messages

		// Отправляем сообщение всем клиентам
		for _, client := range clients {
			client.conn.Write([]byte(message))
		}
	}
}

// Функция для обработки подключения клиента
func handleConnection(client *Client) {
	defer client.conn.Close()

	// Бесконечный цикл для чтения сообщений от клиента
	scanner := bufio.NewScanner(client.conn)
	for scanner.Scan() {
		message := scanner.Text()

		// Формируем сообщение с именем отправителя
		messageWithSender := fmt.Sprintf("[%s]: %s\n", client.name, message)

		// Отправляем сообщение в канал для рассылки другим клиентам
		messages <- messageWithSender
	}
}
