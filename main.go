package mesh

import (
	"encoding/binary"

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
func (controller *Controller) ResetNode(addr uint16) {
	parms := []byte{OpNodeReset}
	parms = append(parms, toByteSlice(addr)...)
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
func (controller *Controller) SendMessage(state byte, addr uint16, appIdx uint16) {
	parms := []byte{OpSendMessage}
	parms = append(parms, state)
	parms = append(parms, toByteSlice(addr)...)
	parms = append(parms, toByteSlice(appIdx)...)
	controller.WriteData(parms)
}

// SendRecallMessage sends a bt mesh scene recall message using the app key at the given index to the given addr
func (controller *Controller) SendRecallMessage(sceneNumber uint16, addr uint16, appIdx uint16) {
	parms := []byte{OpSendRecallMessage}
	parms = append(parms, toByteSlice(sceneNumber)...)
	parms = append(parms, toByteSlice(addr)...)
	parms = append(parms, toByteSlice(appIdx)...)
	controller.WriteData(parms)
}

// SendStoreMessage sends a bt mesh scene store message using the app key at the given index to the given addr
func (controller *Controller) SendStoreMessage(sceneNumber uint16, addr uint16, appIdx uint16) {
	parms := []byte{OpSendStoreMessage}
	parms = append(parms, toByteSlice(sceneNumber)...)
	parms = append(parms, toByteSlice(addr)...)
	parms = append(parms, toByteSlice(appIdx)...)
	controller.WriteData(parms)
}

// SendDeleteMessage sends a bt mesh scene delete message using the app key at the given index to the given addr
func (controller *Controller) SendDeleteMessage(sceneNumber uint16, addr uint16, appIdx uint16) {
	parms := []byte{OpSendDeleteMessage}
	parms = append(parms, toByteSlice(sceneNumber)...)
	parms = append(parms, toByteSlice(addr)...)
	parms = append(parms, toByteSlice(appIdx)...)
	controller.WriteData(parms)
}

// SendBindMessage sends a bt mesh event bind message using the app key at the given index to the given addr
func (controller *Controller) SendBindMessage(sceneNumber uint16, addr uint16, appIdx uint16) {
	parms := []byte{OpSendBindMessage}
	parms = append(parms, toByteSlice(sceneNumber)...)
	parms = append(parms, toByteSlice(addr)...)
	parms = append(parms, toByteSlice(appIdx)...)
	controller.WriteData(parms)
}

// ConfigureNode binds an app key to the node with the given addr
func (controller *Controller) ConfigureNode(addr uint16, appIdx uint16) {
	parms := []byte{OpConfigureNode}
	parms = append(parms, toByteSlice(addr)...)
	parms = append(parms, toByteSlice(appIdx)...)
	controller.WriteData(parms)
}

// ConfigureElem binds an app key to the elem with the given addr
func (controller *Controller) ConfigureElem(groupAddr uint16, nodeAddr uint16, elemAddr uint16, appIdx uint16) {
	parms := []byte{OpConfigureElem}
	parms = append(parms, toByteSlice(groupAddr)...)
	parms = append(parms, toByteSlice(nodeAddr)...)
	parms = append(parms, toByteSlice(elemAddr)...)
	parms = append(parms, toByteSlice(appIdx)...)
	controller.WriteData(parms)
}

// Provision adds a device with the given uuid to the network
func (controller *Controller) Provision(uuid []byte) {
	parms := []byte{OpProvision}
	parms = append(parms, uuid...)
	controller.WriteData(parms)
}

// AddKey generates an app key at the given index
func (controller *Controller) AddKey(appIdx uint16) {
	parms := []byte{OpAddKey}
	parms = append(parms, toByteSlice(appIdx)...)
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

// Only works with unsigned 16 bit numbers
func toByteSlice(imput uint16) []byte {
	bytes := []byte{0x00, 0x00}
	binary.LittleEndian.PutUint16(bytes, imput)
	return bytes
}
