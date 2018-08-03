package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"
)

func main() {
	var port int
	var config string
	var certFile, keyFile string

	flag.IntVar(&port, "--port", 443, "Sets port (defaults to 443)")
	flag.StringVar(&config, "--config", "devproxy.conf", "Specifies configuration file")
	flag.StringVar(&certFile, "--cert", "devproxy.crt", "Specifies TLS certificate")
	flag.StringVar(&keyFile, "--key", "devproxy.key", "Specifies TLS key")
	flag.Parse()

	routes, err := parseConfig(config)
	if err != nil {
		fmt.Printf("Could not parse configurtion: %v\n", err)
		os.Exit(1)
	}

	// Set up proxy
	proxy := newProxy()
	proxy.Routes = routes

	srv, err := newServer(proxy.Handler(), port)
	if err != nil {
		fmt.Printf("Could not start server: %v\n", err)
		os.Exit(1)
	}

	// Set up API
	api, err := newServer(proxy.Api(), port+1)
	if err != nil {
		fmt.Printf("Could not start server: %v\n", err)
		os.Exit(1)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	fmt.Printf("Started proxy server on port %d, API on port %d\n", port, port+1)
	fmt.Printf("Configuration: %s\n", proxy)

	go func() {
		if err := srv.Serve(certFile, keyFile); err != nil {
			fmt.Printf("Could not serve: %v\n", err)
		}
	}()
	go func() {
		if err := api.Serve(certFile, keyFile); err != nil {
			fmt.Printf("Could not serve: %v\n", err)
		}
	}()

	<-stop

	fmt.Printf("Shutting down... ")
	if err := srv.Shutdown(3 * time.Second); err != nil {
		fmt.Printf("Could not gracefully shut down server: %v\n", err)
	}
	fmt.Printf("Done\n")
}
