package lib

import (
	"fmt"
	"log"
	"time"

	"github.com/fatih/color"
)

func LoadBLauncher(clb *LoadBalancer, lbchan chan LoadBalancer) {
	var socks = make(chan interface{})
	go clb.StartCron("@every 10s", lbchan)
	// go clb.ifstart(socks) // can't ioctl and test in macos for now
	go clb.updateSocks(socks)
	go clb.DashboardRender(lbchan)
	// i get old clb state here
	for {
		select {
		case <-socks:
			time.Sleep(10 * time.Second)
			// i get new clb state here
			log.Default().Output(0, color.HiGreenString(fmt.Sprintf("TCPSock: %d \t UDPSock: %d", clb.TCPSock, clb.UDPSock)))
			log.Default().Output(0, color.HiMagentaString(fmt.Sprintf("Use this mac address: %s", generateRandomMAC())))

			go clb.handleTCP()
			go clb.handleUDP()
		case <-time.After(5 * time.Second):
		}
	}
}
