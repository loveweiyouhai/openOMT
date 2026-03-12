package main

import (
	"flag"
	"log"
	"openOMT/internal/server"
)

func main() {
	addr := flag.String("addr", ":8765", "监听地址，例如 :8765 或 127.0.0.1:8765")
	flag.Parse()
	if err := server.Run(*addr); err != nil {
		log.Fatal(err)
	}
}
