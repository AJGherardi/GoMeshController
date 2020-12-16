package mesh

import (
	"encoding/binary"
	"errors"
	"time"

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
	OpState               = 0x13
	OpConfigureElem       = 0x14
	OpConfigureElemStatus = 0x15
	OpSendRecallMessage   = 0x16
	OpSendStoreMessage    = 0x17
	OpSendDeleteMessage   = 0x18
	OpSendBindMessage     = 0x19
	OpEvent               = 0x20
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
func Open() (Controller, error) {
	// Get ctx and defer close func
	ctx := gousb.NewContext()
	// Get device and defer close func
	dev, err := ctx.OpenDeviceWithVIDPID(0x2fe3, 0x0100)
	if err != nil {
		return Controller{}, errors.New("Unable to open controller")
	}
	// Set auto detach from kernel to true
	err = dev.SetAutoDetach(true)
	if err != nil {
		return Controller{}, errors.New("Unable to open controller")
	}
	// Get main config and defer close
	cfg, err := dev.Config(1)
	if err != nil {
		return Controller{}, errors.New("Unable to get config")
	}
	// Get interface 1 and defer close
	intf, err := cfg.Interface(1, 0)
	if err != nil {
		return Controller{}, errors.New("Unable to open interface")
	}
	// Get out and in endpoints
	epIn, err := intf.InEndpoint(2)
	epOut, err := intf.OutEndpoint(1)
	if err != nil {
		return Controller{}, errors.New("Unable to open endpoints")
	}
	// Make struct
	controller := Controller{
		context: ctx,
		device:  dev,
		config:  cfg,
		intf:    intf,
		epIn:    epIn,
		epOut:   epOut,
	}
	return controller, nil
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
	onAddKeyStatus func(appIdx uint16),
	onUnprovisionedBeacon func(uuid []byte),
	onNodeAdded func(addr uint16),
	onState func(addr uint16, state byte),
	onEvent func(addr uint16),
) error {
	for {
		// Read a packet
		buf := make([]byte, controller.epIn.Desc.MaxPacketSize)
		controller.epIn.Read(buf)
		// if err != nil {
		// 	if err != gousb.ErrorOverflow && err != gousb.TransferNoDevice && err != gousb.ErrorIO {
		// 		// return errors.New("Failed to read message")
		// 		log.Fatal(err)
		// 	}
		// 	// If overflow discard message
		// 	continue
		// }
		// Map to provided function
		if buf[0] == OpSetupStatus {
			onSetupStatus()
		}
		if buf[0] == OpAddKeyStatus {
			onAddKeyStatus(binary.LittleEndian.Uint16(buf[1:3]))
		}
		if buf[0] == OpUnprovisionedBeacon {
			onUnprovisionedBeacon(buf[1:17])
		}
		if buf[0] == OpNodeAdded {
			onNodeAdded(binary.LittleEndian.Uint16(buf[1:3]))
		}
		if buf[0] == OpState {
			onState(binary.LittleEndian.Uint16(buf[1:3]), buf[3])
		}
		if buf[0] == OpEvent {
			onEvent(binary.LittleEndian.Uint16(buf[1:3]))
		}
	}
}

// ResetNode Removes the node with the givin addr from the mesh network
func (controller *Controller) ResetNode(addr uint16) error {
	parms := []byte{OpNodeReset}
	parms = append(parms, toByteSlice(addr)...)
	return controller.WriteData(parms)
}

// Reboot reboots the Mesh Controller must be called after reset
func (controller *Controller) Reboot() error {
	return controller.WriteData([]byte{OpReboot})
}

// Reset removes all mesh related items from the Mesh Controller's flash
func (controller *Controller) Reset() error {
	return controller.WriteData([]byte{OpReset})
}

// SendMessage sends a bt mesh message using the app key at the given index to the given addr
func (controller *Controller) SendMessage(state byte, addr uint16, appIdx uint16) error {
	parms := []byte{OpSendMessage}
	parms = append(parms, state)
	parms = append(parms, toByteSlice(addr)...)
	parms = append(parms, toByteSlice(appIdx)...)
	return controller.WriteData(parms)
}

// SendRecallMessage sends a bt mesh scene recall message using the app key at the given index to the given addr
func (controller *Controller) SendRecallMessage(sceneNumber uint16, addr uint16, appIdx uint16) error {
	parms := []byte{OpSendRecallMessage}
	parms = append(parms, toByteSlice(sceneNumber)...)
	parms = append(parms, toByteSlice(addr)...)
	parms = append(parms, toByteSlice(appIdx)...)
	return controller.WriteData(parms)
}

// SendStoreMessage sends a bt mesh scene store message using the app key at the given index to the given addr
func (controller *Controller) SendStoreMessage(sceneNumber uint16, addr uint16, appIdx uint16) error {
	parms := []byte{OpSendStoreMessage}
	parms = append(parms, toByteSlice(sceneNumber)...)
	parms = append(parms, toByteSlice(addr)...)
	parms = append(parms, toByteSlice(appIdx)...)
	return controller.WriteData(parms)
}

// SendDeleteMessage sends a bt mesh scene delete message using the app key at the given index to the given addr
func (controller *Controller) SendDeleteMessage(sceneNumber uint16, addr uint16, appIdx uint16) error {
	parms := []byte{OpSendDeleteMessage}
	parms = append(parms, toByteSlice(sceneNumber)...)
	parms = append(parms, toByteSlice(addr)...)
	parms = append(parms, toByteSlice(appIdx)...)
	return controller.WriteData(parms)
}

// SendBindMessage sends a bt mesh event bind message using the app key at the given index to the given addr
func (controller *Controller) SendBindMessage(sceneNumber uint16, addr uint16, appIdx uint16) error {
	parms := []byte{OpSendBindMessage}
	parms = append(parms, toByteSlice(sceneNumber)...)
	parms = append(parms, toByteSlice(addr)...)
	parms = append(parms, toByteSlice(appIdx)...)
	return controller.WriteData(parms)
}

// ConfigureNode binds an app key to the node with the given addr
func (controller *Controller) ConfigureNode(addr uint16, appIdx uint16) error {
	parms := []byte{OpConfigureNode}
	parms = append(parms, toByteSlice(addr)...)
	parms = append(parms, toByteSlice(appIdx)...)
	return controller.WriteData(parms)
}

// ConfigureElem binds an app key to the elem with the given addr
func (controller *Controller) ConfigureElem(groupAddr uint16, nodeAddr uint16, elemAddr uint16, appIdx uint16) error {
	parms := []byte{OpConfigureElem}
	parms = append(parms, toByteSlice(groupAddr)...)
	parms = append(parms, toByteSlice(nodeAddr)...)
	parms = append(parms, toByteSlice(elemAddr)...)
	parms = append(parms, toByteSlice(appIdx)...)
	return controller.WriteData(parms)
}

// Provision adds a device with the given uuid to the network
func (controller *Controller) Provision(uuid []byte) error {
	parms := []byte{OpProvision}
	parms = append(parms, uuid...)
	return controller.WriteData(parms)
}

// AddKey generates an app key at the given index
func (controller *Controller) AddKey(appIdx uint16) error {
	parms := []byte{OpAddKey}
	parms = append(parms, toByteSlice(appIdx)...)
	return controller.WriteData(parms)
}

// Setup creates a new bt mesh network
func (controller *Controller) Setup() error {
	return controller.WriteData([]byte{OpSetup})
}

// WriteData writes data to the Mesh Controller over usb
func (controller *Controller) WriteData(data []byte) error {
	_, err := controller.epOut.Write(data)
	if err != nil {
		// If write fails retry after a delay
		time.Sleep(200 * time.Millisecond)
		_, err = controller.epOut.Write(data)

		// If write fails again error out
		if err != nil {
			return errors.New("Write failed")
		}
	}
	return nil
}

// Only works with unsigned 16 bit numbers
func toByteSlice(imput uint16) []byte {
	bytes := []byte{0x00, 0x00}
	binary.LittleEndian.PutUint16(bytes, imput)
	return bytes
}
