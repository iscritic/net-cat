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

	flag.StringVar(&port, "port", "8989", "Port to listen on")
	flag.StringVar(&address, "address", "localhost", "Address to listen on")

	flag.Parse()

	lg := nc.NewLogger()

	addr := address + ":" + port

	// Start the server
	lg.InfoLog.Printf("Listening on the port %s...\n", addr)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		lg.ErrorLog.Fatalln("Error starting server:", err)
		return
	}
	defer listener.Close()

}
