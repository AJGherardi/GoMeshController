package main

import (
	"fmt"

	"github.com/google/gousb"
)

// Op codes for the Mesh Controller
const (
	OpSetup               = 0x00
	OpSetupStatus         = 0x01
	OpAddKey              = 0x02
	OpAddKeyStatus        = 0x03
	OpUnprovisionedBeacon = 0x04
	OpProvision           = 0x05
	OpNodeAdded           = 0x06
	OpConfigureNode       = 0x07
	OpConfigureNodeStatus = 0x08
	OpSendMessage         = 0x09
	OpReset               = 0x10
	OpReboot              = 0x11
	OpNodeReset           = 0x12
)

// Controller holds all the needed usb vars to talk to the Mesh Controller
type Controller struct {
	context *gousb.Context
	device  *gousb.Device
	config  *gousb.Config
	intf    *gousb.Interface
	epIn    *gousb.InEndpoint
	epOut   *gousb.OutEndpoint
}

// Open gets the Mesh Controller using usb
func Open() Controller {
	// Get ctx and defer close func
	ctx := gousb.NewContext()
	// Get device and defer close func
	dev, _ := ctx.OpenDeviceWithVIDPID(0x2fe3, 0x0100)
	// Set auto detach from kernel to true
	dev.SetAutoDetach(true)
	// Get main config and defer close
	cfg, _ := dev.Config(1)
	// Get interface 1 and defer close
	intf, _ := cfg.Interface(1, 0)
	// Get out and in endpoints
	epIn, _ := intf.InEndpoint(2)
	epOut, _ := intf.OutEndpoint(1)
	// Make struct
	controller := Controller{
		context: ctx,
		device:  dev,
		config:  cfg,
		intf:    intf,
		epIn:    epIn,
		epOut:   epOut,
	}
	return controller
}

func main() {
	controller := Open()
	go controller.Read(
		func() {
			fmt.Println("setupStatus")
			controller.AddKey([]byte{0x00, 0x00})
		},
		func(appIdx []byte) {
			fmt.Printf("appKeyStatus %x \n", appIdx)
		},
		func(uuid []byte) {
			fmt.Printf("unprovBeacon %x \n", uuid)
			// controller.Provision(uuid)
		},
		func(addr []byte) {
			fmt.Printf("nodeAdded %x \n", addr)
			controller.ConfigureNode(addr, []byte{0x00, 0x00})
		},
	)
	// controller.Reset()
	// time.Sleep(1 * time.Second)
	// controller.Reboot()
	// time.Sleep(1 * time.Second)
	// controller.Setup()
	// controller.Provision([]byte{0x19, 0x8a, 0x1d, 0x0d, 0x7e, 0xd1, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	controller.SendMessage(0x00, []byte{0x00, 0x0a}, []byte{0x00, 0x00})
	// controller.ResetNode([]byte{0x00, 0x08})
	defer controller.Close()
	for {
	}
}

// Close must be called when the Mesh Controller is not needed anymore
func (controller *Controller) Close() {
	controller.intf.Close()
	controller.config.Close()
	controller.device.Close()
	controller.context.Close()
}

// Read calls the provided funcs when a msg from the Mesh Controller is recived
func (controller *Controller) Read(
	onSetupStatus func(),
	onAddKeyStatus func(appIdx []byte),
	onUnprovisionedBeacon func(uuid []byte),
	onNodeAdded func(addr []byte),
) {
	for {
		// Read a packet
		buf := make([]byte, controller.epIn.Desc.MaxPacketSize)
		controller.epIn.Read(buf)
		// Map to provided function
		if buf[0] == OpSetupStatus {
			onSetupStatus()
		}
		if buf[0] == OpAddKeyStatus {
			onAddKeyStatus(buf[1:3])
		}
		if buf[0] == OpUnprovisionedBeacon {
			onUnprovisionedBeacon(buf[1:17])
		}
		if buf[0] == OpNodeAdded {
			onNodeAdded(buf[1:3])
		}
	}
}

// ResetNode Removes the node with the givin addr from the mesh network
func (controller *Controller) ResetNode(addr []byte) {
	parms := []byte{OpNodeReset}
	parms = append(parms, addr...)
	controller.WriteData(parms)
}

// Reboot reboots the Mesh Controller must be called after reset
func (controller *Controller) Reboot() {
	controller.WriteData([]byte{OpReboot})
}

// Reset removes all mesh related items from the Mesh Controller's flash
func (controller *Controller) Reset() {
	controller.WriteData([]byte{OpReset})
}

// SendMessage sends a bt mesh message using the app key at the given index to the given addr
func (controller *Controller) SendMessage(state byte, addr []byte, appIdx []byte) {
	parms := []byte{OpSendMessage}
	parms = append(parms, state)
	parms = append(parms, addr...)
	parms = append(parms, appIdx...)
	controller.WriteData(parms)
}

// ConfigureNode binds an app key to the node with the given addr
func (controller *Controller) ConfigureNode(addr []byte, appIdx []byte) {
	parms := []byte{OpConfigureNode}
	parms = append(parms, addr...)
	parms = append(parms, appIdx...)
	controller.WriteData(parms)
}

// Provision adds a device with the given uuid to the network
func (controller *Controller) Provision(uuid []byte) {
	parms := []byte{OpProvision}
	parms = append(parms, uuid...)
	controller.WriteData(parms)
}

// AddKey generates an app key at the given index
func (controller *Controller) AddKey(appIdx []byte) {
	parms := []byte{OpAddKey}
	parms = append(parms, appIdx...)
	controller.WriteData(parms)
}

// Setup creates a new bt mesh network
func (controller *Controller) Setup() {
	controller.WriteData([]byte{OpSetup})
}

// WriteData writes data to the Mesh Controller over usb
func (controller *Controller) WriteData(data []byte) {
	controller.epOut.Write(data)
}
