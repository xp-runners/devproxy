package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"
)

func main() {
	var apiPort, proxyPort, timeout int
	var configFile, certFile, keyFile string

	flag.IntVar(&proxyPort, "port", 443, "Sets proxy port (defaults to 443)")
	flag.IntVar(&apiPort, "api", 8008, "Sets API port (defaults to 8008)")
	flag.IntVar(&timeout, "timeout", 10, "Sets proxy read and write timeouts (in seconds)")
	flag.StringVar(&configFile, "config", "devproxy.conf", "Specifies configuration file")
	flag.StringVar(&certFile, "cert", "devproxy.crt", "Specifies TLS certificate")
	flag.StringVar(&keyFile, "key", "devproxy.key", "Specifies TLS key")
	flag.Parse()

	config, err := parseConfig(configFile)
	if err != nil {
		fmt.Printf("Could not parse configuration: %v\n", err)
		os.Exit(1)
	}

	// Set up proxy
	proxy := newProxy(proxyPort)
	for _, pair := range config {
		proxy.Proxy(pair.Prefix, pair.Target)
	}

	srv, err := newServer(proxy.Handler(), proxyPort, time.Duration(timeout)*time.Second)
	if err != nil {
		fmt.Printf("Could not start server: %v\n", err)
		os.Exit(1)
	}

	// Set up API
	api, err := newServer(proxy.Api(), apiPort, 3*time.Second)
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
