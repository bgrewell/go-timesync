package main

import (
	"flag"
	"fmt"
	"github.com/bgrewell/go-timesync"
	"os"
	"os/signal"
	"time"

	"github.com/BGrewell/go-conversions"
)

func main() {

	server := flag.Bool("server", false, "Run as a server")
	client := flag.String("client", "", "Run as client and sync with server specified")
	port := flag.Int("port", 9991, "Port to listen on")
	flag.Parse()

	if !*server && *client == "" {
		panic("Must specify either server or client")
	}

	if *server {
		serverMain(*port)
	} else {
		clientMain(*client, *port)
	}

}

func clientMain(host string, port int) {

	err := timesync.EnableSync(host, port)
	if err != nil {
		panic(err)
	}

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt)
		<-ch
		timesync.DisableSync()
		os.Exit(0)
	}()

	for {
		fmt.Printf("offset: %s, delay: %s\n", conversions.ConvertNanosecondsToStringTime(timesync.Offset().Nanoseconds()), conversions.ConvertNanosecondsToStringTime(timesync.Delay().Nanoseconds()))
		time.Sleep(time.Second)
	}
}

func serverMain(port int) {
	var c timesync.Clock
	c = &timesync.SimpleClock{}
	err := c.EnableServer("0.0.0.0", port)
	if err != nil {
		panic(err)
	}

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt)
		<-ch
		c.DisableServer()
		os.Exit(0)
	}()

	for {

	}
}
