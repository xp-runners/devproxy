package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strings"
)

// parseConfig parses the configuration file. Its format is:
//
// /prefix http://proxy.target.host/base
//
// Empty lines and lines beginning with a "#" sign are ignored
func parseConfig(config string) (map[string]*url.URL, error) {
	file, err := os.Open(config)
	if err != nil {
		return nil, fmt.Errorf("Could not open configuration: %v\n", err)
	}
	defer file.Close()

	routes := make(map[string]*url.URL)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}

		tokens := strings.Split(strings.Trim(line, " "), " ")
		if len(tokens) == 2 {
			url, err := url.Parse(tokens[1])
			if err != nil {
				return nil, fmt.Errorf("Malformed URL %s: %v\n", tokens[1], err)
			}
			routes[tokens[0]] = url
		}
	}

	return routes, scanner.Err()
}
