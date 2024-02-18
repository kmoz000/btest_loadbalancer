package lib

// import (
// 	"fmt"
// 	"runtime"

// 	"github.com/songgao/water"
// )

// // Tuntap represents a TUN/TAP interface
// type Tuntap struct {
// 	ifce *water.Interface
// }

// // NewTuntap creates a new TUN/TAP interface
// func (t *Tuntap) create(name string) error {
// 	var _type water.DeviceType
// 	var err error
// 	var prefix string
// 	if runtime.GOOS == "windows" {
// 		_type = water.TAP
// 		prefix = "utap"
// 	} else {
// 		_type = water.TUN
// 		prefix = "utun"
// 	}
// 	config := water.Config{
// 		DeviceType: _type,
// 	}
// 	config.Name = fmt.Sprintf("%s%s", prefix, name)
// 	if t.ifce, err = water.New(config); err != nil {
// 		return err
// 	}
// 	if err = setInterfaceState(config.Name); err != nil {
// 		return err
// 	}
// 	return nil
// }

// // Read reads data from the TUN/TAP interface
// func (t *Tuntap) Read(buffer []byte) (int, error) {
// 	return t.ifce.Read(buffer)
// }

// // Write writes data to the TUN/TAP interface
// func (t *Tuntap) Write(buffer []byte) (int, error) {
// 	return t.ifce.Write(buffer)
// }

// // Close closes the TUN/TAP interface
// func (t *Tuntap) Close() error {
// 	if t.ifce != nil {
// 		return t.ifce.Close()
// 	} else {
// 		return fmt.Errorf("no interface found")
// 	}
// }
