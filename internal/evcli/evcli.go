package evcli

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/cenkalti/rpc2"
)

type Car struct {
	VIN         string
	Locx        int
	Locy        int
	Speed       int
	DriveStatus string
	IsPacer     bool
}

func StartCLI() {
	fmt.Print("\033[H\033[2J")
	scanner := bufio.NewScanner(os.Stdin)
	flag.Usage = func() {
		fmt.Printf("Use below CLI commands:\n")
		fmt.Printf("show all\n")
		fmt.Printf("show vin <vin>\n")
		fmt.Printf("============================\n")
		flag.PrintDefaults()
	}
	flag.Usage()

	for scanner.Scan() {
		fmt.Printf("============================\n")
		line := scanner.Text()
		if line == "exit" {
			os.Exit(0)
		}
		s := strings.Split(line, " ")
		if len(s) < 2 {
			flag.Usage()
			continue
		}
		switch s[0] {
		case "show":
			if s[1] == "all" {
				ExecuteCmd(line)
			} else if s[1] == "vin" && len(s) == 3 {
				ExecuteCmd(line)
			} else {
				flag.Usage()
				continue
			}
		default:
			flag.Usage()
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error in reading input:", err)
	}
}

func ExecuteCmd(cmd string) {

	// Connecting  with local server
	cli, err := net.Dial("tcp", "localhost:8009")
	if err != nil {
		fmt.Println("EV Server is not running")
		return
	}
	// Create new intance and then query for data
	var reply []*Car
	clt := rpc2.NewClient(cli)
	defer clt.Close()

	go clt.Run()
	// Issue CLI command
	err = clt.Call("CLI", cmd, &reply)
	if err != nil {
		fmt.Println("Failed to get data error:", err)
		return
	}
	l := len(reply)
	if l == 0 {
		fmt.Println("No vehicle data found")
		return
	}
	// Vehicle data is comming as slice(array) of all vehicles
	fmt.Println("::Vehicles Information::")
	for i := 0; i < l; i++ {
		fmt.Printf("VIN:%s, IsPacer:%t, Coordinates:(%d,%d), Speed:%d\n", reply[i].VIN, reply[i].IsPacer, reply[i].Locx, reply[i].Locy, reply[i].Speed)
	}
}
