package main

import (
	"encoding/binary"
	"log"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func TestLoadBalancer_All(t *testing.T) {
	// server1 := lib.Server{Host: "host1", Bind: []lib.PortRange{{Port: 8080}}}
	// server2 := lib.Server{Host: "host2", Bind: []lib.PortRange{{Port: 8081}}}
	// server3 := lib.Server{Host: "host3", Bind: []lib.PortRange{{Port: 8082}}}

	// lb := lib.LoadBalancer{
	// 	AliveNodes: map[string]lib.Server{
	// 		"server1": server1,
	// 		"server2": server2,
	// 	},
	// 	DeadNodes: map[string]lib.Server{
	// 		"server3": server3,
	// 	},
	// }

	// result := lb.All()

	// expected := map[string]lib.Server{
	// 	"server1": server1,
	// 	"server2": server2,
	// 	"server3": server3,
	// }

	// if !reflect.DeepEqual(result, expected) {
	// 	t.Errorf("Expected %v, but got %v", expected, result)
	// }
	// host := "64.20.217.236"
	// for _, port := range []int{2000, 2001} {
	// 	timeout := time.Second
	// 	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, strconv.Itoa(port)), timeout)
	// 	if err != nil {
	// 		fmt.Println("Connecting error:", err)
	// 	}
	// 	if conn != nil {
	// 		defer conn.Close()
	// 		fmt.Println("Opened", net.JoinHostPort(host, strconv.Itoa(port)))
	// 	}
	// }
	handle, err := pcap.OpenLive("any", 1600, true, pcap.BlockForever)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	for packet := range packetSource.Packets() {
		// Check if it's a UDP packet
		udpLayer := packet.TransportLayer()
		if udpLayer != nil && udpLayer.LayerType() == layers.LayerTypeUDP {
			// Print information about the UDP packet
			gotpacket := udpLayer.TransportFlow()
			distPrt := binary.BigEndian.Uint16(gotpacket.Dst().Raw())
			srcPrt := binary.BigEndian.Uint16(gotpacket.Src().Raw())
			log.Default().Println("Received UDP packet from:", srcPrt, " to:", distPrt, " lne:", len(gotpacket.Dst().Raw()))
			if distPrt > 2000 && distPrt <= 3000 {
				log.Default().Printf("Layer: %v", udpLayer.LayerType())
			}
			// log.Default().Println("Port:", udpLayer.TransportFlow())
			// log.Default().Printf("Payload: %v\n", udpLayer.LayerPayload())
		}
	}
}
