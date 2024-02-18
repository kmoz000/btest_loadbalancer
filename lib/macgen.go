package lib

import (
	"fmt"
	"math/rand"
	"time"
)

// Define a struct to represent a MAC address range
type MACRange struct {
	Prefix string
}

// List of MAC address ranges
var macRanges = []MACRange{
	{Prefix: "08:55:31"},
	{Prefix: "18:FD:74"},
	{Prefix: "2C:C8:1B"},
	{Prefix: "48:8F:5A"},
	{Prefix: "48:A9:8A"},
	{Prefix: "4C:5E:0C"},
	{Prefix: "64:D1:54"},
	{Prefix: "6C:3B:6B"},
	{Prefix: "74:4D:28"},
	{Prefix: "78:9A:18"},
	{Prefix: "B8:69:F4"},
	{Prefix: "C4:AD:34"},
	{Prefix: "CC:2D:E0"},
	{Prefix: "D4:CA:6D"},
	{Prefix: "DC:2C:6E"},
	{Prefix: "E4:8D:8C"},
}

func generateRandomMAC() string {
	rand.Seed(time.Now().UnixNano())

	// Randomly select a MAC address range
	selectedRange := macRanges[rand.Intn(len(macRanges))]

	mac := make([]byte, 6)
	// Parse the selected prefix into the first three bytes of the MAC address
	for i := 0; i < 3; i++ {
		val, _ := fmt.Sscanf(selectedRange.Prefix[i*3:i*3+2], "%02X", &mac[i])
		if val != 1 {
			panic("Invalid MAC address prefix")
		}
	}

	// Generate the remaining three bytes
	for i := 3; i < 6; i++ {
		mac[i] = byte(rand.Intn(256))
	}

	// Format the MAC address
	return fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X", mac[0], mac[1], mac[2], mac[3], mac[4], mac[5])
}
