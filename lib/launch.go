package lib

import (
	"fmt"
	"log"
	"time"

	"github.com/fatih/color"
)

func LoadBLauncher(clb *LoadBalancer, lbchan chan LoadBalancer) {
	if clb.TryLock() {
		defer clb.Unlock()
	}
	var socks = make(chan interface{})
	var NotifyrSocks = make(chan bool, 1)
	go clb.StartCron("@every 10s", lbchan)
	// go clb.ifstart(socks) // can't ioctl and test in macos for now
	go clb.updateSocks(socks)
	go clb.DashboardRender(lbchan)
	// I get old clb state here
	for {
		select {
		case <-socks:
			time.Sleep(10 * time.Second)
			// i get new clb state here
			log.Default().Output(0, color.HiGreenString(fmt.Sprintf("TCPConn: %s", clb.TCPSock.Addr().String())))
			log.Default().Output(0, color.HiMagentaString(fmt.Sprintf("Use this mac address: %s", generateRandomMAC())))

			go clb.handleTCP(NotifyrSocks)
			// go clb.handleUDP(NotifyrSocks)
		case <-NotifyrSocks:
			go clb.updateSocks(socks)
		case <-time.After(5 * time.Second):
		}
	}
}
