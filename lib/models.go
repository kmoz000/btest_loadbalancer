package lib

import (
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"net/netip"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

const (
	FROMhost = "0.0.0.0"
)

// Server represents the server configuration in the YAML file
type Server struct {
	Host    string      `yaml:"host"`
	Bind    []PortRange `yaml:"bind"`
	Pinging int64       `yaml:"pinging,omitempty"`
	Peers   []Peer      `yaml:"peers,omitempty"`
}
type Peer struct {
	connected_at time.Time
	ip           string
	action       Action
}
type Server_ struct {
	Host string `yaml:"host"`
	Bind string `yaml:"bind"`
}

// LoadBalancer represents a simple TCP load balancer
type LoadBalancer struct {
	AliveNodes map[string]Server
	DeadNodes  map[string]Server
	// Interface  Tuntap
	TCPSock *net.TCPListener
	UDPSock *net.UDPConn
	sync.Mutex
}

func (lb *LoadBalancer) deletePeerByIP(serverName, peerIP string) {
	if lb.TryLock() {
		defer lb.Unlock()
	}
	if server, ok := lb.AliveNodes[serverName]; ok {
		var updatedPeers []Peer
		for _, peer := range server.Peers {
			if peer.ip != peerIP {
				updatedPeers = append(updatedPeers, peer)
			}
		}
		server.Peers = updatedPeers
		lb.AliveNodes[serverName] = server
	}
}

// HealthCheckJob is a cron job that checks the health of server nodes
func (lb *LoadBalancer) HealthCheckJob() {
	if lb.TryLock() {
		defer lb.Unlock()
	}
	allServers := lb.All()
	var wg sync.WaitGroup
	resultChan := make(chan struct {
		name    string
		server  Server
		isAlive bool
		ping    int64
	}, len(allServers))
	for name, server := range allServers {
		wg.Add(1)
		go func(name string, server Server) {
			defer wg.Done()
			isAlive, ping := isServerAlive(&server)
			resultChan <- struct {
				name    string
				server  Server
				isAlive bool
				ping    int64
			}{name, server, isAlive, ping}
		}(name, server)
	}
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	for result := range resultChan {
		if !result.isAlive {
			result.server.Pinging = result.ping
			lb.DeadNodes[result.name] = result.server
			delete(lb.AliveNodes, result.name)
		} else {
			result.server.Pinging = 0
			lb.AliveNodes[result.name] = result.server
			delete(lb.DeadNodes, result.name)
		}
	}
	// logs need to be deleted after
	log.Default().Output(1, color.GreenString(fmt.Sprintln("Alive Nodes:", lb.AliveNodes)))
	log.Default().Output(1, color.RedString(fmt.Sprintln("Dead Nodes:", lb.DeadNodes)))
}
func (lb *LoadBalancer) All() map[string]Server {
	allServers := make(map[string]Server)
	for key, value := range lb.AliveNodes {
		allServers[key] = value
	}
	for key, value := range lb.DeadNodes {
		allServers[key] = value
	}
	return allServers
}

// func (lb *LoadBalancer) ifstart(sockc chan interface{}) {
// 	// Create new listeners based on AliveNodes
// 	var err error

// 	if err = lb.Interface.create("19"); err != nil {
// 		logCreationError(lb.Interface.ifce.IsTAP(), err)

// 		if errors.Is(err, syscall.Errno(syscall.EPERM)) {
// 			if permissionErr := lb.requestPermission(); permissionErr != nil {
// 				log.Default().Output(0, "not permitted!")
// 			}
// 		}
// 	} else {
// 		defer func() {
// 			syscall.Close(lb.TCPSock)
// 			syscall.Close(lb.UDPSock)
// 		}()

// 		addr := syscall.SockaddrInet4{
// 			Port: 0,                   // Use 0 to listen on all ports
// 			Addr: [4]byte{0, 0, 0, 0}, // Listen on all available interfaces
// 		}

// 		lb.TCPSock, err = syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
// 		if err != nil {
// 			log.Default().Output(0, err.Error())
// 			return
// 		}

// 		lb.UDPSock, err = syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
// 		if err != nil {
// 			log.Default().Output(0, err.Error())
// 			return
// 		}
// 		if _, err = syscall.BpfInterface(lb.TCPSock, lb.Interface.ifce.Name()); err != nil {
// 			log.Default().Output(0, err.Error())
// 			return
// 		}
// 		if _, err = syscall.BpfInterface(lb.UDPSock, lb.Interface.ifce.Name()); err != nil {
// 			log.Default().Output(0, err.Error())
// 			return
// 		}
// 		err = syscall.Bind(lb.TCPSock, &addr)
// 		if err != nil {
// 			log.Default().Output(0, err.Error())
// 			return
// 		}

// 		err = syscall.Bind(lb.UDPSock, &addr)
// 		if err != nil {
// 			log.Default().Output(0, err.Error())
// 			return
// 		}

//			sockc <- make(chan interface{})
//		}
//	}
func (lb *LoadBalancer) handleTCP(notify chan bool) {
	if lb.TryLock() {
		defer lb.Unlock()
	}
	for lb.TCPSock != nil {
		// Accept new connections
		name, _server := lb.serverPool()
		if conn, err := lb.TCPSock.Accept(); err == nil && _server != nil {
			action := make(chan Action, 100)
			portchan := make(chan uint16, 100)
			go lb.handleTCPConn(conn, _server, action, portchan, name)
			go lb.handlePeer(conn, &action, &portchan, _server, name)
			go handleUDPConn(conn.RemoteAddr(), _server, portchan)
		}
	}
	notify <- true
}

// func (lb *LoadBalancer) handleUDP(notify chan bool) {
// 	buffer := make([]byte, 1501)

//		for lb.UDPSock != nil {
//			// Read from UDP connection
//			var err error
//			n, sender, err := lb.UDPSock.ReadFromUDP(buffer)
//			if err != nil {
//				log.Default().Println("Error reading from UDP:", err)
//				continue
//			}
//			var clientCon *net.UDPConn
//			if clientCon, err = net.DialUDP("udp4", nil, sender); err != nil {
//				continue
//			}
//			name, _server := lb.serverPool()
//			// if _, ok := lb.AliveNodes[name]; ok && _server != nil {
//			// 	_server.Peers = append(lb.AliveNodes[name].Peers, Peer{
//			// 		ip:           addr.String(),
//			// 		connected_at: time.Now(),
//			// 	})
//			// 	lb.AliveNodes[name] = *_server
//			// }
//			if _server != nil && n > 0 {
//				go handleUDPConn(clientCon, sender, _server, name, buffer[:n])
//			}
//		}
//		notify <- true
//	}
func (lb *LoadBalancer) handleTCPConn(us net.Conn, server *Server, action chan Action, portchan chan uint16, name string) {
	dicon := make(chan string, 1)
	ds, err := net.DialTCP("tcp",
		// net.TCPAddrFromAddrPort(netip.MustParseAddrPort(us.RemoteAddr().String())),
		nil,
		net.TCPAddrFromAddrPort(netip.MustParseAddrPort(net.JoinHostPort(server.Host, strconv.Itoa(server.Bind[0].Port)))))
	if err != nil {
		us.Close()
		log.Printf("failed to dial %s: %s", server.Host, err)
		return
	}
	host, _, err := net.SplitHostPort(us.RemoteAddr().String())
	if err != nil {
		log.Default().Printf("failed to split host and port: %s", err)
		return
	}
	go copyConn(ds, us, nil, action, portchan, "")
	go copyConn(us, ds, &dicon, nil, portchan, host)
	go lb.handleDisconnectedPeers(&dicon, portchan, action, server, name)
}

func (lb *LoadBalancer) handleDisconnectedPeers(dicon *chan string, portchan chan uint16, action chan Action, server *Server, name string) {
	for {
		disconnectedIP := <-*dicon
		lb.deletePeerByIP(name, disconnectedIP)
		// close(portchan)
		// close(action)
	}
}
func (lb *LoadBalancer) handlePeer(conn net.Conn, action *chan Action, portchan *chan uint16, server *Server, name string) {
	for {
		PeerAction := <-*action
		if _, ok := lb.AliveNodes[name]; ok && PeerAction.TxSize > 0 {
			log.Default().Printf("cmd: %v, %s", PeerAction, name)
			host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
			if err != nil {
				log.Default().Printf("failed to split host and port: %s", err)
				return
			}
			server.Peers = append(lb.AliveNodes[name].Peers, Peer{
				ip:           host,
				connected_at: time.Now(),
				action:       PeerAction,
			})
			lb.AliveNodes[name] = *server
		}
	}
}

// serverPool retrieves the server with lower pinging and fewer peers.
func (lb *LoadBalancer) serverPool() (string, *Server) {
	var selectedServer *Server
	var selectedName string
	minPinging := int64(math.MinInt64)
	minPeers := math.MaxInt64
	if len(lb.AliveNodes) == 0 {
		return "", nil
	}
	for name, server := range lb.AliveNodes {
		// Compare pinging
		if len(server.Peers) < minPeers {
			selectedServer = &server
			selectedName = name
			minPeers = len(server.Peers)
		} else if len(server.Peers) == minPeers {
			// If pinging is equal, compare peers count
			if server.Pinging < minPinging {
				selectedServer = &server
				selectedName = name
				minPinging = server.Pinging
			}
		}
	}

	return selectedName, selectedServer
}
func (lb *LoadBalancer) updateSocks(socks chan interface{}) {
	if lb.TryLock() {
		defer lb.Unlock()
	}
	var err error
	tcpaddr := net.TCPAddr{
		Port: 2000,
		IP:   net.IPv4(0, 0, 0, 0),
	}
	// udpaddr := net.UDPAddr{
	// 	Port: 0,
	// 	IP:   net.IPv4(0, 0, 0, 0),
	// }
	if lb.TCPSock, err = net.ListenTCP("tcp", &tcpaddr); err != nil {
		log.Default().Println("Listner: ", err.Error())
		return
	}
	// if lb.UDPSock, err = net.ListenPacket("udp4", &udpaddr); err != nil {
	// 	log.Default().Output(0, err.Error())
	// 	return
	// }
	socks <- make(chan interface{})
}

// func (lb *LoadBalancer) requestPermission() error {
// 	validate := func(input string) error {
// 		if input != "yes" {
// 			return fmt.Errorf("permission denied")
// 		}
// 		return nil
// 	}

// 	prompt := promptui.Prompt{
// 		Label:     "Do you want to grant permission? (Type 'yes' to confirm)",
// 		Validate:  validate,
// 		IsConfirm: true,
// 	}

// 	result, err := prompt.Run()
// 	if err != nil {
// 		return err
// 	}

// 	fmt.Printf("You entered %s\n", result)
// 	return nil
// }

func (lb *LoadBalancer) StartCron(expr string, lbchan chan LoadBalancer) {
	c := cron.New()
	// Schedule the health check job to run every 10 seconds
	_, err := c.AddFunc(expr, lb.HealthCheckJob)
	if err != nil {
		log.Fatal("Error adding health check job to cron:", err)
	}
	// Start the cron scheduler in a separate goroutine
	go c.Start()
	lbchan <- *lb
}

// PortRange represents a range of ports for a specific protocol
type PortRange struct {
	From     int    `yaml:"from,omitempty"`
	To       int    `yaml:"to,omitempty"`
	IsRange  bool   `yaml:"isrange,omitempty"`
	Port     int    `yaml:"port,omitempty"`
	Protocol string `yaml:"protocol"`
}

// parseBindString parses the bind string and returns a list of PortRange structs
func parseBindString(bind string) ([]PortRange, error) {
	var portRanges []PortRange

	// Match patterns like "2001-2256/tcp" or "2000/tcp"
	pattern := regexp.MustCompile(`(\d+)(?:-(\d+))?/([a-zA-Z]+)`)
	matches := pattern.FindAllStringSubmatch(bind, -1)
	isrange := false
	for _, match := range matches {
		start, err := strconv.Atoi(match[1])
		if err != nil {
			return nil, fmt.Errorf("error parsing port range start: %v", err)
		}
		// If there is no match[2], it means there is no port range
		var end int
		if match[2] != "" {
			end, err = strconv.Atoi(match[2])
			isrange = true
			if err != nil {
				return nil, fmt.Errorf("error parsing port range end: %v", err)
			}
		} else {
			// If there is no range, use start as end
			end = start
		}

		protocol := match[3]
		var portRange PortRange
		if isrange {
			portRange = PortRange{
				From:     start,
				To:       end,
				IsRange:  isrange,
				Protocol: protocol,
			}
		} else {
			portRange = PortRange{
				Port:     start,
				IsRange:  isrange,
				Protocol: protocol,
			}
		}
		portRanges = append(portRanges, portRange)
	}

	return portRanges, nil
}

func (s *Server) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var _server Server_
	if err := unmarshal(&_server); err == nil {
		// If unmarshaling as a string is successful, parse it as a single port or range
		portRanges, err := parseBindString(_server.Bind)
		if err != nil {
			return fmt.Errorf("error parsing bind string: %v", err)
		}
		s.Host = _server.Host
		s.Bind = portRanges
		return nil
	}

	// If unmarshaling as a string fails, try unmarshaling as a list
	var server Server
	if err := unmarshal(&server); err != nil {
		return fmt.Errorf("error unmarshaling bind: %v", err)
	}
	// fmt.Printf("%v\n", bindList)
	s.Host = server.Host
	s.Bind = server.Bind
	return nil
}

// dashboard dyn template Rendering
func (lb *LoadBalancer) DashboardRender(lbchan chan LoadBalancer) {
	if lb.TryLock() {
		defer lb.Unlock()
	}
	// gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	// Marshal the LoadBalancer instance to a JSON string
	router.LoadHTMLGlob("templates/index*")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.htm", gin.H{"live": lb.AliveNodes, "dead": lb.DeadNodes})
	})
	log.Default().Println(color.HiMagentaString(fmt.Sprintf("DASHBOARD RUNNING ON PORT: %d", 8080)))
	router.Run(":8080")
}

// Config represents the overall configuration structure
type Config struct {
	Servers map[string]Server `yaml:"servers"`
}

// Action struct to represent an action
type Action struct {
	Proto         string
	Direction     string
	Random        bool
	TCPConnCount  int
	TxSize        uint16
	Unknown       uint32
	RemoteTxSpeed uint32
}
