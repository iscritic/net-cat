package main

import (
	"bufio"
	"net"
	"os"

	nc "nc/pkg/logger"
)

var messages = make(chan string)
var clients []net.Conn

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

		// Добавляем клиента в список
		clients = append(clients, conn)

		// Обработка клиента в отдельной горутине
		go handleConnection(conn)
	}
}

// Функция для обработки входящих сообщений от клиентов
func broadcastMessages() {
	for {
		// Получаем сообщение из канала
		message := <-messages

		// Отправляем сообщение всем клиентам
		for _, client := range clients {
			client.Write([]byte(message))
		}
	}
}

// Функция для обработки подключения клиента
func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Приветствие нового клиента
	conn.Write([]byte("Welcome to the chat!\n"))

	// Бесконечный цикл для чтения сообщений от клиента
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		message := scanner.Text()

		if message == "" {
			continue
		} else {
			message = message + "\n"
		}

		// Отправляем сообщение в канал для рассылки другим клиентам
		messages <- message
	}
}
