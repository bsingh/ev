package evclient

import (
	"crypto/rand"
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/cenkalti/rpc2"
)

// DriveStatus -  "parked", "driving", "reverse"
// Commands - "honk", "toggle headlights", "toggle door lock"

type StatusReply string
type CommandReply int

var (
	// shutdown channel to handle ctr-c
	shutdown = make(chan os.Signal, 1)
)

type Register struct {
	VIN string
}

type Car struct {
	VIN         string
	Locx        int
	Locy        int
	Speed       int
	DriveStatus string
	IsPacer     bool
}

type EVClient struct {
	me     Car
	server string
	con    net.Conn
}

func (e *EVClient) Connect() error {
	fmt.Println("EV Client connecting with server")
	for {
		// Will keep trying connecting every 5 seconds till successful connection
		con, err := net.Dial("tcp", e.server)
		if err == nil {
			e.con = con
			return nil
		}
		fmt.Println("Failed to connect with EV Server retry after 5 seconds:", err)
		select {
		case <-shutdown:
			fmt.Println("EV Client shutting down")
			os.Exit(0)
		case <-time.After(5 * time.Second):
			fmt.Println("EV Client re-connecting")

		}
	}
}

func (e *EVClient) Run() error {

	//Instantiate new client
	clt := rpc2.NewClient(e.con)
	clt.Handle("DriveStatus", func(client *rpc2.Client, status int, reply *StatusReply) error {
		*reply = StatusReply(e.me.DriveStatus)
		return nil
	})
	clt.Handle("Command", func(client *rpc2.Client, cmd string, reply *CommandReply) error {
		*reply = CommandReply(0)
		switch cmd {
		case "honk":
			fmt.Println("Car is honking now!")
		case "toggle headlights":
			fmt.Println("Car doing headlight toggle!")
		case "toggle door lock":
			fmt.Println("Car doing door lock toggle!")
		default:
			fmt.Println("Car found unknown command", cmd)
			*reply = CommandReply(-1)
		}
		return nil
	})

	go clt.Run()

	var reply int
	fmt.Println("Registering with VIN", e.me.VIN)
	err := clt.Call("Register", e.me.VIN, &reply)
	if err != nil {
		fmt.Println("Registering error:", err)
		return err
	}
	fmt.Println("Registering completed")

	counter := 0
	for {
		// Using fake counter to reset every 10th time
		switch counter {
		case 0: // park it!
			e.me.Locx = 30 + counter
			e.me.Locy = 20 - counter
			e.me.Speed = 0
			e.me.DriveStatus = "parked"
		case -1: // reverse
			e.me.Locx = 11 - counter
			e.me.Locy = 20 + counter
			e.me.Speed = 10 - counter
			e.me.DriveStatus = "reverse"

		default: // driving
			e.me.Locx = 30 - counter
			e.me.Locy = 20 + counter
			e.me.Speed = 50 + counter
			e.me.DriveStatus = "driving"

		}
		fmt.Printf("EV Sending Vehicle Stats VIN:%s Location:%d,%d Speed:%d DriverStatus:%s\n", e.me.VIN, e.me.Locx, e.me.Locy, e.me.Speed, e.me.DriveStatus)

		err = clt.Call("UpdateStats", e.me, &reply)
		if err != nil {
			fmt.Println("UpdateStats error:", err)
			return err
		}
		select {
		case <-shutdown:
			fmt.Println("EV Vehicles hutting down")
			fmt.Println("Unregistering VIN", e.me.VIN)
			_ = clt.Call("UnRegister", e.me.VIN, &reply)
			clt.Close()
			return nil
		case <-time.After(5 * time.Second):
		}

		if counter == 0 {
			counter = -1 // make reverse after park!
		} else if counter > 20 {
			counter = 0 // stop it after going too fast!
		} else {
			counter += 2
		}

	}

}

func StartClient(srv string) {

	signal.Notify(shutdown, os.Interrupt)

	evc := EVClient{server: srv}
	vin := RandomVIN()
	evc.me = Car{VIN: vin}

	for {
		evc.Connect()
		err := evc.Run()
		if err == nil {
			return
		}
		select {
		case <-shutdown:
			fmt.Println("EV Vehicle shutting down")
			return
		case <-time.After(5 * time.Second):
			fmt.Println("EV Vehicle retrying after 5  seconds")
		}
	}
}

func RandomVIN() string {
	b := make([]byte, 5)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	s := fmt.Sprintf("J%X", b)
	return s
}
