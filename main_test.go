package main

import (
	"fmt"
	"net"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/kmoz000/btest_loadbalancer/lib"
)

func TestLoadBalancer_All(t *testing.T) {
	server1 := lib.Server{Host: "host1", Bind: []lib.PortRange{{Port: 8080}}}
	server2 := lib.Server{Host: "host2", Bind: []lib.PortRange{{Port: 8081}}}
	server3 := lib.Server{Host: "host3", Bind: []lib.PortRange{{Port: 8082}}}

	lb := lib.LoadBalancer{
		AliveNodes: map[string]lib.Server{
			"server1": server1,
			"server2": server2,
		},
		DeadNodes: map[string]lib.Server{
			"server3": server3,
		},
	}

	result := lb.All()

	expected := map[string]lib.Server{
		"server1": server1,
		"server2": server2,
		"server3": server3,
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, but got %v", expected, result)
	}
	host := "64.20.217.236"
	for _, port := range []int{2000, 2001} {
		timeout := time.Second
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, strconv.Itoa(port)), timeout)
		if err != nil {
			fmt.Println("Connecting error:", err)
		}
		if conn != nil {
			defer conn.Close()
			fmt.Println("Opened", net.JoinHostPort(host, strconv.Itoa(port)))
		}
	}
}
