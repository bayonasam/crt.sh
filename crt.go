package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type CrtshResult struct {
	NameValue string `json:"name_value"`
}

func spinner(delay time.Duration) {
	for {
		for _, r := range `-\|/` {
			fmt.Printf("\r%c", r)
			time.Sleep(delay)
		}
	}
}

func main() {
	fmt.Println(`
                    __                    
    _____   _____  / /_      ____ _  ____ 
   / ___/  / ___/ / __/     / __  / / __ \
  / /__   / /    / /_   _  / /_/ / / /_/ /
  \___/  /_/     \__/  (_) \__, /  \\___/ 
                          /____/ @TaurusOmar_
	`)

	// Define command line flags
	domain := flag.String("d", "", "Domain to scan")
	output := flag.String("o", "", "Output file path")
	flag.Parse()

	// Check if domain is provided
	if *domain == "" {
		fmt.Printf("Use: %s -d <domain> [-o <output file path>]\n", os.Args[0])
		os.Exit(1)
	}

	// Default output directory if not provided
	if *output == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		*output = filepath.Join(homeDir, "result_directory", fmt.Sprintf("%s.crt.txt", *domain))
	}

	crtURL := fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", *domain)

	fmt.Printf("Scanning for domain: %s...\n", *domain)

	go spinner(100 * time.Millisecond)

	resp, err := http.Get(crtURL)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	var crtshResults []CrtshResult
	err = json.Unmarshal(body, &crtshResults)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	subdomains := make(map[string]struct{})
	for _, result := range crtshResults {
		subdomain := strings.TrimPrefix(result.NameValue, "*.")
		subdomains[subdomain] = struct{}{}
	}

	uniqueSubdomains := make([]string, 0, len(subdomains))
	for subdomain := range subdomains {
		uniqueSubdomains = append(uniqueSubdomains, subdomain)
	}

	sort.Strings(uniqueSubdomains)

	outputDir := filepath.Dir(*output)
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err = os.MkdirAll(outputDir, 0755)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
	}

	err = ioutil.WriteFile(*output, []byte(strings.Join(uniqueSubdomains, "\n")), 0644)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("\rScan completed. Results saved in %s\n", *output)
}
