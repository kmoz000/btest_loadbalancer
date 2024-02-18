package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kmoz000/btest_loadbalancer/lib"
	"gopkg.in/yaml.v2"
)

var lbchan = make(chan lib.LoadBalancer, 1)

func main() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	// Define a command-line flag for the YAML file path
	filePath := flag.String("f", "", "Path to the YAML configuration file")
	flag.Parse()

	// If -f flag is not provided, check if the config file is provided as the first argument
	if *filePath == "" && len(os.Args) > 1 {
		*filePath = os.Args[1]
	}

	// Check if the file path is provided
	if *filePath == "" {
		log.Fatal("Please provide the path to the YAML configuration file using the -f flag or as the first argument.")
	}

	// Read the YAML file
	yamlFile, err := os.ReadFile(*filePath)
	if err != nil {
		log.Fatalf("Error reading YAML file: %v", err)
	}

	// Parse the YAML content into a Config struct
	var config lib.Config
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		yamlError, ok := err.(*yaml.TypeError)
		if ok {
			log.Fatalf("Error in YAML schema at line %d: %s", len(yamlError.Errors), yamlError)
		} else {
			log.Fatalf("Error parsing YAML content: %v", err)
		}
	}
	clb := &lib.LoadBalancer{
		AliveNodes: config.Servers,
		DeadNodes:  make(map[string]lib.Server),
		// Interface:  lib.Tuntap{},
	}
	go lib.LoadBLauncher(clb, lbchan)
	// Call the LoadBLauncher function with the updated configuration
	i := 0
	for {
		select {
		case <-lbchan:
			log.Default().Output(1, "Loadbalancer started ✅️")
			i++
		case <-sigCh:
			// if err := clb.Interface.Close(); err != nil {
			// 	log.Default().Output(0, err.Error())
			// }
			os.Exit(0)
		}
	}
	// Wait for a signal to terminate the program (e.g., via OS interrupt signal)
}
