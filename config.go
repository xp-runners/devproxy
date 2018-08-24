package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strings"
)

type pair struct {
	Prefix string
	Target *url.URL
}

// parseConfig parses the configuration file. Its format is:
//
// /prefix http://proxy.target.host/base
//
// Empty lines and lines beginning with a "#" sign are ignored
func parseConfig(config string) ([]pair, error) {
	file, err := os.Open(config)
	if err != nil {
		return nil, fmt.Errorf("Could not open configuration: %v\n", err)
	}
	defer file.Close()

	pairs := make([]pair, 0)
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
			pairs = append(pairs, pair{tokens[0], url})
		}
	}

	return pairs, scanner.Err()
}
