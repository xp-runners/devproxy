package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"
)

func main() {
	var apiPort, proxyPort int
	var config, certFile, keyFile string

	flag.IntVar(&proxyPort, "port", 443, "Sets proxy port (defaults to 443)")
	flag.IntVar(&apiPort, "api", 8008, "Sets API port (defaults to 8008)")
	flag.StringVar(&config, "config", "devproxy.conf", "Specifies configuration file")
	flag.StringVar(&certFile, "cert", "devproxy.crt", "Specifies TLS certificate")
	flag.StringVar(&keyFile, "key", "devproxy.key", "Specifies TLS key")
	flag.Parse()

	routes, err := parseConfig(config)
	if err != nil {
		fmt.Printf("Could not parse configurtion: %v\n", err)
		os.Exit(1)
	}

	// Set up proxy
	proxy := newProxy(proxyPort)
	proxy.Routes = routes

	srv, err := newServer(proxy.Handler(), proxyPort)
	if err != nil {
		fmt.Printf("Could not start server: %v\n", err)
		os.Exit(1)
	}

	// Set up API
	api, err := newServer(proxy.Api(), apiPort)
	if err != nil {
		fmt.Printf("Could not start server: %v\n", err)
		os.Exit(1)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	fmt.Printf("Started proxy server on port %d (https), API on port %d (http)\n", proxyPort, apiPort)
	fmt.Printf("Configuration: %s\n", proxy)

	go func() {
		if err := srv.ServeTLS(certFile, keyFile); err != nil {
			fmt.Printf("Could not serve: %v\n", err)
		}
	}()
	go func() {
		if err := api.Serve(); err != nil {
			fmt.Printf("Could not serve: %v\n", err)
		}
	}()

	s := <-stop

	fmt.Printf("Shutting down on %v... ", s)
	if err := srv.Shutdown(3 * time.Second); err != nil {
		fmt.Printf("Could not gracefully shut down server: %v\n", err)
	}
	fmt.Printf("Done\n")
}
