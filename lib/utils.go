package lib

import (
	"io"
	"log"
	"math"
	"net"
	"strconv"
	"sync"
	"time"
)

// isServerAlive checks if a server is alive by attempting a connection to its bound ports
func isServerAlive(server *Server) (bool, int64) {
	resultChan := make(chan int64, 1)
	if len(server.Bind) > 0 {
		portRange := server.Bind[0]
		if portRange.IsRange {
			ports := []int{portRange.From, portRange.To}
			var wg sync.WaitGroup
			resultChan := make(chan int64, len(ports))
			for _, port := range ports {
				wg.Add(1)
				go isPortOpenWg(server.Host, portRange.Protocol, port, &wg, resultChan)
			}
			min := math.MaxInt64
			for result := range resultChan {
				if result == 0 {
					return false, 0
				}
				if int(result) < min {
					min = int(result)
				}
			}
			return true, int64(min)
		} else {
			var wg sync.WaitGroup

			wg.Add(1)
			go isPortOpenWg(server.Host, portRange.Protocol, portRange.Port, &wg, resultChan)
			for {
				select {
				case ping := <-resultChan:
					if ping != 0 {
						return true, ping
					}
					return false, 0
				case <-time.After(1 * time.Second):
				}
			}
		}
	} else {
		return false, 0
	}
}

// func logCreationError(isTAP bool, err error) {
// 	var networkType string
// 	if isTAP {
// 		networkType = "TAP"
// 	} else {
// 		networkType = "TUN"
// 	}

// 	log.Default().Output(0, fmt.Sprintf("can't create  (%s): %s", networkType, err.Error()))
// }
// func setInterfaceState(ifaceName string) error {
// 	lnk, err := netlink.LinkByName(ifaceName)
// 	if err != nil {
// 		return err
// 	}

// 	ipConfig := &netlink.Addr{IPNet: &net.IPNet{
// 		IP:   net.ParseIP("19.0.0.1"),
// 		Mask: net.CIDRMask(24, 32),
// 	}}

// 	if err = netlink.AddrAdd(lnk, ipConfig); err != nil {
// 		// log.Default().Output(0, err.Error())
// 		return err
// 	}
// 	return nil
// }

// isPortOpen checks if a port on a server is open by attempting a TCP connection
func isPortOpen(host string, protocol string, port int) bool {
	// log.Default().Output(1, fmt.Sprintf("Checking %s for access to port %d", host, port))
	// Validate the protocol
	if protocol != "tcp" && protocol != "udp" {
		log.Printf("Unsupported protocol: %s for port %d\n", protocol, port)
		return false
	}
	address := net.JoinHostPort(host, strconv.Itoa(port))
	// Attempt a connection based on the protocol
	conn, err := net.DialTimeout(protocol, address, time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()

	return true
}
func isPortOpenWg(host string, protocol string, port int, wg *sync.WaitGroup, resultChan chan int64) {
	defer wg.Done()

	// Validate the protocol
	if protocol != "tcp" && protocol != "udp" {
		log.Printf("Unsupported protocol: %s for port %d\n", protocol, port)
		resultChan <- time.Duration(0).Milliseconds()
		return
	}

	address := net.JoinHostPort(host, strconv.Itoa(port))

	// Attempt a connection based on the protocol
	startTime := time.Now()
	conn, err := net.DialTimeout(protocol, address, time.Second)
	if err != nil {
		resultChan <- time.Duration(0).Milliseconds()
		return
	}
	defer conn.Close()

	// Calculate the round-trip time (ping)
	duration := time.Since(startTime).Milliseconds()
	resultChan <- duration
}

// func handleUDP(ifce *water.Interface) {
// 	// Use the provided UDP socket
// 	// You might need to configure the socket for listening and handling UDP traffic
// 	// For simplicity, this example uses the existing socket and just prints received datagrams
// 	var frame ethernet.Frame

//		for {
//			frame.Resize(1500)
//			n, err := ifce.Read([]byte(frame))
//			if err != nil {
//				log.Fatal(err)
//			}
//			frame = frame[:n]
//			log.Printf("Dst: %s\n", frame.Destination())
//			log.Printf("Src: %s\n", frame.Source())
//			log.Printf("Ethertype: % x\n", frame.Ethertype())
//			log.Printf("Payload: % x\n", frame.Payload())
//			time.Sleep(5 * time.Second)
//		}
//	}
func copyConn(wc io.WriteCloser, r io.Reader, dicon *chan string, name string) {
	defer wc.Close()
	io.Copy(wc, r)
	if name != "" {
		*dicon <- name
	}
}
func copyUDPData(dst *net.UDPConn, srcAddr *net.UDPAddr, data []byte) {
	_, err := dst.WriteToUDP(data, srcAddr)
	if err != nil {
		log.Default().Printf("Error writing UDP data: %s", err)
		return
	}
}
