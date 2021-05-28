package main

import (
	"fmt"
	"log"
	"os"

	evcli "github.com/bsingh/ev/internal/evcli"
	evcln "github.com/bsingh/ev/internal/evclient"
	evsrv "github.com/bsingh/ev/internal/evserver"
)

// Main Launcher App
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Launch ", os.Args[0], " with one of these commands server, client, cli")
		log.Fatal(1)
	}
	command := os.Args[1]
	fmt.Println("running command", command)
	switch command {
	case "server":
		evsrv.StartServer()
	case "client":
		server := "localhost:8009"
		if len(os.Args) == 3 {
			server = os.Args[2]
		}
		evcln.StartClient(server)
	case "cli":
		evcli.StartCLI()
	default:
		fmt.Println("Launch ", os.Args[0], " with one of these commands server, client or cli")
		log.Fatal(1)
	}
}
