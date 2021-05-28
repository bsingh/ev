package evserver

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

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

var (
	// Shutdown channel to handle ctr-c
	shutdown = make(chan os.Signal, 1)
)

type EVServer struct {
	srv      *rpc2.Server   // Server connection
	vehicles map[string]Car // Map to store vehivle info with vin as key
	pacer    string         // Used to store pace car
	mu       sync.Mutex     // Mutex to synchronize data access
}

// Registering vehicle
func (e *EVServer) Register(vin string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	fmt.Println("EV Server registering new vehicle VIN", vin)
	if _, ok := e.vehicles[vin]; ok {
		fmt.Println("EV Server error re-registering VIN", vin)
		return errors.New("Re-registering")
	}
	e.vehicles[vin] = Car{VIN: vin}
	// Handle very first time case
	if e.pacer == "" {
		go e.assignPaceVehicle()
	}
	return nil
}

// UnRegistering vehicle
func (e *EVServer) UnRegister(vin string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	fmt.Println("EV Server unregistering vehicle VIN", vin)
	v, ok := e.vehicles[vin]
	if !ok {
		return errors.New("Un-registering unknown vehicle")
	}

	if v.IsPacer {
		fmt.Println("EV Server pacer vehicle disconnected so assign new one")
		e.pacer = ""
		go e.assignPaceVehicle()
	}
	delete(e.vehicles, vin)
	return nil
}

// Update Stats for given vehicle
func (e *EVServer) UpdateStats(car *Car) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	v, ok := e.vehicles[car.VIN]
	if !ok {
		fmt.Println("EV Server Updating unknown vehicle")
		return errors.New("Updating unknown vehicle")
	}
	// Updating member with new values
	v.DriveStatus = car.DriveStatus
	v.Locx = car.Locx
	v.Locy = car.Locy
	v.Speed = car.Speed

	e.vehicles[car.VIN] = v

	fmt.Printf("EV Server UpdateStats VIN:%s IsPacer:%t Location:%d,%d Speed:%d DriverStatus:%s\n", v.VIN, v.IsPacer, v.Locx, v.Locy, v.Speed, v.DriveStatus)
	return nil
}

// Serve CLI command from CLI
func (e *EVServer) CLIExec(cli *string) []*Car {
	e.mu.Lock()
	defer e.mu.Unlock()

	res := []*Car{}
	s := strings.Split(*cli, " ")
	if s[1] == "all" {
		// Pull all cars and send in batch
		for _, v := range e.vehicles {
			c := v
			res = append(res, &c)
		}
	} else {
		// This must be for given vin
		v, ok := e.vehicles[s[2]]
		if !ok {
			fmt.Println("EV Server this vin not found", s[2])
			return res
		}
		res = append(res, &v)
	}
	//for i := 0; i < len(res); i++ {
	//	fmt.Printf("VIN:%s,IsPacer:%t, Coordinates:(%d,%d), Speed:%d\n", res[i].VIN, res[i].IsPacer, res[i].Locx, res[i].Locy, res[i].Speed)
	//}
	return res
}

// This function will be called by main thread
func (e *EVServer) assignPaceVehicle() {
	e.mu.Lock()
	defer e.mu.Unlock()
	// Assgin first value in map as pace vehicle
	for _, v := range e.vehicles {
		fmt.Println("EV Server assign new vehicle as pacer VIN", v.VIN)
		v.IsPacer = true
		e.vehicles[v.VIN] = v
		// Ensure flag is set in global object
		e.pacer = v.VIN
		break
	}
}

func StartServer() {
	signal.Notify(shutdown, os.Interrupt)
	evs := EVServer{}

	s := rpc2.NewServer()
	v := make(map[string]Car)
	evs.srv = s
	evs.vehicles = v

	// Handler for registration
	evs.srv.Handle("Register", func(client *rpc2.Client, vin *string, reply *int) error {
		evs.Register(*vin)
		*reply = 0
		return nil
	})
	// Handler for un-registration
	evs.srv.Handle("UnRegister", func(client *rpc2.Client, vin *string, reply *int) error {
		evs.UnRegister(*vin)
		*reply = 0
		return nil
	})
	// Handler for update stats
	evs.srv.Handle("UpdateStats", func(client *rpc2.Client, car *Car, reply *int) error {
		// Lets check Drive Status when speed zero
		if car.Speed == 0 {
			var status string
			client.Call("DriveStatus", 0, &status)
			fmt.Println("EV Server DriveStatus on zero speed", status) // parked
		}
		if car.DriveStatus == "reverse" {
			var status int
			client.Call("Command", "honk", &status)
			fmt.Println("EV Server sent honk on reverse!")
		}
		evs.UpdateStats(car)

		*reply = 0
		return nil
	})
	// Handler for service CLI commands
	evs.srv.Handle("CLI", func(client *rpc2.Client, cli *string, reply *[]*Car) error {
		fmt.Println("EV Server CLI", *cli)
		*reply = evs.CLIExec(cli)
		fmt.Println("EV Server CLI returned")
		return nil
	})

	// More logic can be added to mange connections
	evs.srv.OnDisconnect(func(client *rpc2.Client) {
		return
	})

	tcpAddr, err := net.ResolveTCPAddr("tcp", ":8009")
	checkError(err)
	fmt.Println("EV Server tcpAddr", tcpAddr)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	fmt.Println("EV Server waiting for new connection")
	go evs.srv.Accept(listener)

	for {
		select {
		case <-shutdown:
			fmt.Println("EV Server shutting down")
			os.Exit(0)
		case <-time.After(5 * time.Second):
			// Future use for house keeping things like to check if any bad(unupdated)
			// exists in memory
		}
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}
