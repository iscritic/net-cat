package main

import (
	"flag"
	nc "nc/pkg/logger"
	"net"
)

func main() {

	var (
		port    string
		address string
	)

	flag.String("port", "8989", "Port to listen on")
	flag.String("address", "localhost", "Address to listen on")

	flag.Parse()

	lg := nc.NewLogger()

	// Start the server
	listener, err := net.Listen("tcp", address+":"+port)
	if err != nil {
		lg.ErrorLog.Fatalln("Error starting server:", err)
		return
	}
	defer listener.Close()

	lg.InfoLog.Printf("Listening on the port %s...\n", address+":"+port)

}
