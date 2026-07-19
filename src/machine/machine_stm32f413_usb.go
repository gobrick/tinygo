//go:build stm32f413

package machine

import (
	"device/stm32"
	"machine/usb"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

const NumberOfUSBEndpoints = 8

const (
	otgFSBase       = uintptr(0x50000000)
	otgDeviceBase   = otgFSBase + 0x800
	otgInEPBase     = otgFSBase + 0x900
	otgOutEPBase    = otgFSBase + 0xb00
	otgPowerBase    = otgFSBase + 0xe00
	otgFIFOBase     = otgFSBase + 0x1000
	otgEndpointStep = uintptr(0x20)
	otgFIFOStep     = uintptr(0x1000)
)

const (
	regGOTGCTL    = otgFSBase + 0x000
	regGAHBCFG    = otgFSBase + 0x008
	regGUSBCFG    = otgFSBase + 0x00c
	regGRSTCTL    = otgFSBase + 0x010
	regGINTSTS    = otgFSBase + 0x014
	regGINTMSK    = otgFSBase + 0x018
	regGRXSTSP    = otgFSBase + 0x020
	regGRXFSIZ    = otgFSBase + 0x024
	regDIEPTXF0   = otgFSBase + 0x028
	regGCCFG      = otgFSBase + 0x038
	regDIEPTXF1   = otgFSBase + 0x104
	regDCFG       = otgDeviceBase + 0x000
	regDCTL       = otgDeviceBase + 0x004
	regDIEPMSK    = otgDeviceBase + 0x010
	regDOEPMSK    = otgDeviceBase + 0x014
	regDAINT      = otgDeviceBase + 0x018
	regDAINTMSK   = otgDeviceBase + 0x01c
	regDIEPEMPMSK = otgDeviceBase + 0x034
	regPCGCCTL    = otgPowerBase
)

const (
	regEPCTL   = uintptr(0x00)
	regEPINT   = uintptr(0x08)
	regEPTSIZ  = uintptr(0x10)
	regDTXFSTS = uintptr(0x18)
)

const (
	gotgctlBVALOEN   = uint32(1 << 6)
	gotgctlBVALOVAL  = uint32(1 << 7)
	gahbcfgGINT      = uint32(1 << 0)
	gusbcfgPHYSEL    = uint32(1 << 6)
	gusbcfgTRDTMask  = uint32(0xf << 10)
	gusbcfgTRDT16MHz = uint32(0xd << 10)
	gusbcfgFDMOD     = uint32(1 << 30)
	grstctlCSRST     = uint32(1 << 0)
	grstctlRXFFLSH   = uint32(1 << 4)
	grstctlTXFFLSH   = uint32(1 << 5)
	grstctlTXFAll    = uint32(0x10 << 6)
	grstctlAHBIDL    = uint32(1 << 31)
	gccfgPWRDWN      = uint32(1 << 16)
	gccfgVBDEN       = uint32(1 << 21)
)

const (
	gintRXFLVL  = uint32(1 << 4)
	gintUSBRST  = uint32(1 << 12)
	gintENUMDNE = uint32(1 << 13)
	gintIEPINT  = uint32(1 << 18)
	gintOEPINT  = uint32(1 << 19)
)

const (
	dcfgDSPDFullSpeed = uint32(3)
	dcfgDADMask       = uint32(0x7f << 4)
	dctlSDIS          = uint32(1 << 1)
	epctlUSBAEP       = uint32(1 << 15)
	epctlEPTYPPos     = uint32(18)
	epctlSTALL        = uint32(1 << 21)
	epctlTXFNUMPos    = uint32(22)
	epctlCNAK         = uint32(1 << 26)
	epctlSNAK         = uint32(1 << 27)
	epctlSD0PID       = uint32(1 << 28)
	epctlEPENA        = uint32(1 << 31)
	epintXFRC         = uint32(1 << 0)
	epintTOC          = uint32(1 << 3)
	epintSTUP         = uint32(1 << 3)
	epintTXFE         = uint32(1 << 7)
	eptsizXFRMask     = uint32(0x7ffff)
	eptsizPKTCNTMask  = uint32(0x3ff << 19)
	eptsizPKTCNT1     = uint32(1 << 19)
	eptsizSTUPCNT3    = uint32(3 << 29)
	dtxfstsWordsMask  = uint32(0xffff)
)

const (
	rxPacketData          = uint8(2)
	rxPacketOUTComplete   = uint8(3)
	rxPacketSetupComplete = uint8(4)
	rxPacketSetupData     = uint8(6)
	usbEndpointCount      = uint32(4)
	usbPacketSize         = uint32(64)
	usbIRQEvents          = uint8(8)
	usbPLLM               = uint32(8)
	usbPLLN               = uint32(96)
	usbPLLQ               = uint32(4)
	usbWaitLimit          = uint32(2_000_000)
	usbControlPollLimit   = uint32(4_000_000)
	usbDisconnectPeriods  = uint8(8)
	usbINPacketCountMax   = uint32(1023)
)

// ck48SourcePLLI2SQ is RM0430 DCKCFGR2.CK48MSEL, not the generated PLLSAI name.
const ck48SourcePLLI2SQ = uint32(1 << 27)

const (
	usbRXFIFO  = uint16(128)
	usbTX0FIFO = uint16(32)
	usbTX1FIFO = uint16(16)
	usbTX2FIFO = uint16(16)
	usbTX3FIFO = uint16(64)
)

// This one-line policy must match the schematic; 7a assumes PA9 VBUS sense is absent.
const usbUseVBUSSense = false

type stm32USBReceiveStatus struct {
	ep     uint8
	count  uint16
	pid    uint8
	packet uint8
}

type stm32USBInTransfer struct {
	data   []byte
	offset int
	zlp    bool
	active bool
}

var (
	endPoints = []uint32{
		usb.CONTROL_ENDPOINT:  usb.ENDPOINT_TYPE_CONTROL,
		usb.CDC_ENDPOINT_ACM:  usb.ENDPOINT_TYPE_INTERRUPT | usb.EndpointIn,
		usb.CDC_ENDPOINT_OUT:  usb.ENDPOINT_TYPE_BULK | usb.EndpointOut,
		usb.CDC_ENDPOINT_IN:   usb.ENDPOINT_TYPE_BULK | usb.EndpointIn,
		usb.HID_ENDPOINT_IN:   usb.ENDPOINT_TYPE_DISABLE,
		usb.HID_ENDPOINT_OUT:  usb.ENDPOINT_TYPE_DISABLE,
		usb.MIDI_ENDPOINT_IN:  usb.ENDPOINT_TYPE_DISABLE,
		usb.MIDI_ENDPOINT_OUT: usb.ENDPOINT_TYPE_DISABLE,
	}
	stm32USBInterrupt      interrupt.Interrupt
	stm32USBSetup          [8]byte
	stm32USBSetupReady     bool
	stm32USBOutCount       [NumberOfUSBEndpoints]uint8
	stm32USBEP0Data        []byte
	stm32USBEP0Offset      int
	stm32USBStatusIN       bool
	stm32USBStatusOUT      bool
	stm32USBAddress        uint8
	stm32USBAddressPending bool
	stm32USBEP3Transfer    stm32USBInTransfer
	stm32USBEP3Aborted     bool
)

func usbRegister(address uintptr) *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Pointer(address))
}

func usbInRegister(ep uint32, offset uintptr) *volatile.Register32 {
	return usbRegister(otgInEPBase + uintptr(ep)*otgEndpointStep + offset)
}

func usbOutRegister(ep uint32, offset uintptr) *volatile.Register32 {
	return usbRegister(otgOutEPBase + uintptr(ep)*otgEndpointStep + offset)
}

func usbTransmitFIFO(ep uint32) *volatile.Register32 {
	return usbRegister(otgFIFOBase + uintptr(ep)*otgFIFOStep)
}

func usbReceiveFIFO() *volatile.Register32 {
	return usbRegister(otgFIFOBase)
}

func decodeSTM32USBReceiveStatus(value uint32) stm32USBReceiveStatus {
	return stm32USBReceiveStatus{
		ep:     uint8(value & 0xf),
		count:  uint16((value >> 4) & 0x7ff),
		pid:    uint8((value >> 15) & 0x3),
		packet: uint8((value >> 17) & 0xf),
	}
}

// Configure starts the F413 USB device controller.
func (dev *USBDevice) Configure(_ UARTConfig) {
	if dev.initcomplete {
		return
	}
	establishUSBDisconnected()
	if !startUSBController() {
		disconnectUSB()
		return
	}
	dev.initcomplete = true
	enableOTGInterrupt()
	usbRegister(regDCTL).ClearBits(dctlSDIS)
}

func startUSBController() bool {
	configureUSBPins()
	if !configureUSBClock48() || !resetOTGFS() {
		return false
	}
	usbRegister(regDCTL).SetBits(dctlSDIS)
	if !waitUSBDisconnect() {
		return false
	}
	return configureOTGDevice()
}

func establishUSBDisconnected() {
	resetOTGPeripheral()
	usbRegister(regDCTL).SetBits(dctlSDIS)
	usbRegister(regGCCFG).ClearBits(gccfgPWRDWN)
}

func disconnectUSB() {
	usbRegister(regDCTL).SetBits(dctlSDIS)
	usbRegister(regGAHBCFG).ClearBits(gahbcfgGINT)
	usbRegister(regGCCFG).ClearBits(gccfgPWRDWN)
}

func configureUSBPins() {
	USBCDC_DM_PIN.ConfigureAltFunc(PinConfig{Mode: PinModePWMOutput}, AF10_OTG_FS)
	USBCDC_DP_PIN.ConfigureAltFunc(PinConfig{Mode: PinModePWMOutput}, AF10_OTG_FS)
	USBCDC_DM_PIN.configurePushPullNoPullHighSpeed()
	USBCDC_DP_PIN.configurePushPullNoPullHighSpeed()
}

func configureUSBClock48() bool {
	stm32.RCC.CR.ClearBits(stm32.RCC_CR_PLLI2SON | stm32.RCC_CR_PLLON)
	if !waitRCCClear(stm32.RCC_CR_PLLI2SRDY | stm32.RCC_CR_PLLRDY) {
		return false
	}
	stm32.RCC.CR.ClearBits(stm32.RCC_CR_HSEON)
	if !waitRCCClear(stm32.RCC_CR_HSERDY) {
		return false
	}
	stm32.RCC.CR.ClearBits(stm32.RCC_CR_HSEBYP)
	stm32.RCC.CR.SetBits(stm32.RCC_CR_HSEON)
	if !waitRCCSet(stm32.RCC_CR_HSERDY) {
		return false
	}
	stm32.RCC.PLLCFGR.ReplaceBits(
		usbPLLM|1<<stm32.RCC_PLLCFGR_PLLSRC_Pos,
		stm32.RCC_PLLCFGR_PLLM_Msk|stm32.RCC_PLLCFGR_PLLSRC_Msk, 0)
	stm32.RCC.PLLI2SCFGR.ReplaceBits(
		usbPLLM|usbPLLN<<stm32.RCC_PLLI2SCFGR_PLLI2SN_Pos|usbPLLQ<<stm32.RCC_PLLI2SCFGR_PLLI2SQ_Pos,
		stm32.RCC_PLLI2SCFGR_PLLI2SM_Msk|stm32.RCC_PLLI2SCFGR_PLLI2SN_Msk|
			stm32.RCC_PLLI2SCFGR_PLLI2SQ_Msk|stm32.RCC_PLLI2SCFGR_PLLI2SSRC_Msk, 0)
	stm32.RCC.DCKCFGR2.SetBits(ck48SourcePLLI2SQ)
	stm32.RCC.CR.SetBits(stm32.RCC_CR_PLLI2SON)
	if waitRCCSet(stm32.RCC_CR_PLLI2SRDY) {
		return true
	}
	stm32.RCC.CR.ClearBits(stm32.RCC_CR_PLLI2SON)
	return false
}

func waitRCCSet(mask uint32) bool {
	for count := uint32(0); count < usbWaitLimit; count++ {
		if stm32.RCC.CR.HasBits(mask) {
			return true
		}
	}
	return false
}

func waitRCCClear(mask uint32) bool {
	for count := uint32(0); count < usbWaitLimit; count++ {
		if stm32.RCC.CR.Get()&mask == 0 {
			return true
		}
	}
	return false
}

func resetOTGFS() bool {
	resetOTGPeripheral()
	usbRegister(regGUSBCFG).SetBits(gusbcfgPHYSEL)
	if !waitUSBRegisterSet(regGRSTCTL, grstctlAHBIDL) {
		return false
	}
	usbRegister(regGRSTCTL).SetBits(grstctlCSRST)
	if !waitUSBRegisterClear(regGRSTCTL, grstctlCSRST) {
		return false
	}
	usbRegister(regGUSBCFG).SetBits(gusbcfgFDMOD)
	return true
}

func resetOTGPeripheral() {
	stm32.RCC.AHB2ENR.SetBits(stm32.RCC_AHB2ENR_OTGFSEN)
	stm32.RCC.AHB2RSTR.SetBits(stm32.RCC_AHB2RSTR_OTGFSRST)
	stm32.RCC.AHB2RSTR.ClearBits(stm32.RCC_AHB2RSTR_OTGFSRST)
}

func waitUSBRegisterSet(address uintptr, mask uint32) bool {
	for count := uint32(0); count < usbWaitLimit; count++ {
		if usbRegister(address).HasBits(mask) {
			return true
		}
	}
	return false
}

func waitUSBRegisterClear(address uintptr, mask uint32) bool {
	for count := uint32(0); count < usbWaitLimit; count++ {
		if usbRegister(address).Get()&mask == 0 {
			return true
		}
	}
	return false
}

func waitUSBDisconnect() bool {
	last := TIM2.Count()
	periods := uint8(0)
	for count := uint32(0); count < usbWaitLimit; count++ {
		current := TIM2.Count()
		if current < last {
			periods++
			if periods == usbDisconnectPeriods {
				return true
			}
		}
		last = current
	}
	return false
}

func configureOTGDevice() bool {
	if !waitUSBDeviceMode() {
		return false
	}
	configureUSBVBUS()
	usbRegister(regPCGCCTL).Set(0)
	usbRegister(regGCCFG).SetBits(gccfgPWRDWN)
	usbRegister(regGUSBCFG).ReplaceBits(gusbcfgTRDT16MHz, gusbcfgTRDTMask, 0)
	usbRegister(regGAHBCFG).Set(0)
	usbRegister(regDCFG).Set(dcfgDSPDFullSpeed)
	usbRegister(regDCTL).Set(dctlSDIS)
	clearUSBEndpoints()
	configureUSBFIFOs()
	if !flushUSBFIFOs() {
		return false
	}
	configureUSBInterruptMasks()
	initEndpoint(0, usb.ENDPOINT_TYPE_CONTROL)
	armEP0Setup()
	return true
}

func waitUSBDeviceMode() bool {
	for count := uint32(0); count < usbWaitLimit; count++ {
		if usbRegister(regGINTSTS).Get()&1 == 0 {
			return true
		}
	}
	return false
}

func configureUSBVBUS() {
	if usbUseVBUSSense {
		usbRegister(regGCCFG).SetBits(gccfgVBDEN)
		usbRegister(regGOTGCTL).ClearBits(gotgctlBVALOEN | gotgctlBVALOVAL)
		return
	}
	usbRegister(regGCCFG).ClearBits(gccfgVBDEN)
	usbRegister(regGOTGCTL).SetBits(gotgctlBVALOEN | gotgctlBVALOVAL)
}

func clearUSBEndpoints() {
	usbRegister(regDIEPMSK).Set(0)
	usbRegister(regDOEPMSK).Set(0)
	usbRegister(regDAINTMSK).Set(0)
	usbRegister(regDIEPEMPMSK).Set(0)
	for ep := uint32(0); ep < usbEndpointCount; ep++ {
		usbInRegister(ep, regEPCTL).Set(epctlSNAK)
		usbInRegister(ep, regEPTSIZ).Set(0)
		usbInRegister(ep, regEPINT).Set(0xfb7f)
		usbOutRegister(ep, regEPCTL).Set(epctlSNAK)
		usbOutRegister(ep, regEPTSIZ).Set(0)
		usbOutRegister(ep, regEPINT).Set(0xfb7f)
	}
}

func configureUSBFIFOs() {
	start := usbRXFIFO
	usbRegister(regGRXFSIZ).Set(uint32(usbRXFIFO))
	usbRegister(regDIEPTXF0).Set(fifoValue(start, usbTX0FIFO))
	start += usbTX0FIFO
	usbRegister(regDIEPTXF1).Set(fifoValue(start, usbTX1FIFO))
	start += usbTX1FIFO
	usbRegister(regDIEPTXF1 + 4).Set(fifoValue(start, usbTX2FIFO))
	start += usbTX2FIFO
	usbRegister(regDIEPTXF1 + 8).Set(fifoValue(start, usbTX3FIFO))
}

func fifoValue(start, depth uint16) uint32 {
	return uint32(depth)<<16 | uint32(start)
}

func flushUSBFIFOs() bool {
	return flushUSBTXFIFOs() && flushUSBRXFIFO()
}

func flushUSBTXFIFOs() bool {
	if !waitUSBRegisterSet(regGRSTCTL, grstctlAHBIDL) {
		return false
	}
	usbRegister(regGRSTCTL).Set(grstctlTXFFLSH | grstctlTXFAll)
	if !waitUSBRegisterClear(regGRSTCTL, grstctlTXFFLSH) {
		return false
	}
	return true
}

func flushUSBRXFIFO() bool {
	usbRegister(regGRSTCTL).Set(grstctlRXFFLSH)
	return waitUSBRegisterClear(regGRSTCTL, grstctlRXFFLSH)
}

func configureUSBInterruptMasks() {
	usbRegister(regGINTMSK).Set(0)
	usbRegister(regGINTSTS).Set(0xbfffffff)
	usbRegister(regDIEPMSK).Set(epintXFRC | epintTOC)
	usbRegister(regDOEPMSK).Set(epintXFRC | epintSTUP)
	usbRegister(regDAINTMSK).Set(1 | 1<<16)
	usbRegister(regGINTMSK).Set(gintRXFLVL | gintUSBRST | gintENUMDNE | gintIEPINT | gintOEPINT)
}

func enableOTGInterrupt() {
	stm32USBInterrupt = interrupt.New(stm32.IRQ_OTG_FS, handleSTM32USBInterrupt)
	stm32USBInterrupt.SetPriority(0xe0)
	stm32USBInterrupt.Enable()
	usbRegister(regGAHBCFG).SetBits(gahbcfgGINT)
}

func handleSTM32USBInterrupt(interrupt.Interrupt) {
	serviceUSB(usbIRQEvents)
}

func serviceUSB(maxEvents uint8) {
	for events := uint8(0); events < maxEvents; events++ {
		pending := usbRegister(regGINTSTS).Get() & usbRegister(regGINTMSK).Get()
		switch {
		case pending&gintUSBRST != 0:
			handleUSBReset()
		case pending&gintENUMDNE != 0:
			handleUSBEnumerationDone()
		case pending&gintRXFLVL != 0:
			handleUSBRXStatus()
		case pending&gintIEPINT != 0:
			handleUSBInComplete()
		case pending&gintOEPINT != 0:
			handleUSBOutComplete()
		default:
			return
		}
	}
}

func handleUSBReset() {
	usbRegister(regGINTSTS).Set(gintUSBRST)
	stm32USBEP0Data = nil
	stm32USBEP0Offset = 0
	stm32USBAddressPending = false
	stm32USBSetupReady = false
	stm32USBStatusIN = false
	stm32USBStatusOUT = false
	stm32USBEP3Aborted = stm32USBEP3Aborted || stm32USBEP3Transfer.active
	stm32USBEP3Transfer = stm32USBInTransfer{}
	usbRegister(regDCFG).ClearBits(dcfgDADMask)
	clearUSBEndpoints()
	if !flushUSBTXFIFOs() {
		disconnectUSB()
		return
	}
	configureUSBInterruptMasks()
	initEndpoint(0, usb.ENDPOINT_TYPE_CONTROL)
	armEP0Setup()
	usbConfiguration = 0
	USBDev.InitEndpointComplete = false
}

func handleUSBEnumerationDone() {
	usbRegister(regGINTSTS).Set(gintENUMDNE)
	usbRegister(regDCFG).ReplaceBits(dcfgDSPDFullSpeed, 3, 0)
	initEndpoint(0, usb.ENDPOINT_TYPE_CONTROL)
	armEP0Setup()
}

func handleUSBRXStatus() {
	status := decodeSTM32USBReceiveStatus(usbRegister(regGRXSTSP).Get())
	handleUSBReceiveStatus(status)
}

func handleUSBReceiveStatus(status stm32USBReceiveStatus) {
	switch status.packet {
	case rxPacketSetupData:
		readUSBRXFIFO(stm32USBSetup[:], status.count)
		stm32USBSetupReady = true
	case rxPacketData:
		handleUSBOUTData(status)
	case rxPacketOUTComplete, rxPacketSetupComplete:
		return
	default:
		drainUSBRXFIFO(status.count)
	}
}

func handleUSBOUTData(status stm32USBReceiveStatus) {
	if int(status.ep) >= len(udd_ep_out_cache_buffer) {
		drainUSBRXFIFO(status.count)
		return
	}
	count := status.count
	if count > uint16(len(udd_ep_out_cache_buffer[status.ep])) {
		count = uint16(len(udd_ep_out_cache_buffer[status.ep]))
	}
	readUSBRXFIFO(udd_ep_out_cache_buffer[status.ep][:count], status.count)
	stm32USBOutCount[status.ep] = uint8(count)
}

func readUSBRXFIFO(destination []byte, count uint16) {
	for offset := uint16(0); offset < count; offset += 4 {
		word := usbReceiveFIFO().Get()
		for index := uint16(0); index < 4 && offset+index < count; index++ {
			position := int(offset + index)
			if position < len(destination) {
				destination[position] = byte(word >> (8 * index))
			}
		}
	}
}

func drainUSBRXFIFO(count uint16) {
	readUSBRXFIFO(nil, count)
}

func handleUSBSetup() {
	setup := usb.NewSetup(stm32USBSetup[:])
	stm32USBStatusIN = setup.BmRequestType&usb.REQUEST_DIRECTION == 0
	stm32USBStatusOUT = false
	ok := false
	if setup.BmRequestType&usb.REQUEST_TYPE == usb.REQUEST_STANDARD {
		ok = handleStandardSetup(setup)
	} else if setup.WIndex < uint16(len(usbSetupHandler)) && usbSetupHandler[setup.WIndex] != nil {
		ok = usbSetupHandler[setup.WIndex](setup)
	}
	if !ok {
		USBDev.SetStallEPIn(0)
		USBDev.SetStallEPOut(0)
		armEP0Setup()
	}
}

func handleUSBInComplete() {
	pending := usbRegister(regDAINT).Get() & usbRegister(regDAINTMSK).Get() & 0xffff
	for ep := uint32(0); ep < usbEndpointCount; ep++ {
		if pending&(1<<ep) == 0 {
			continue
		}
		flags := usbInRegister(ep, regEPINT).Get() & usbInInterruptMask(ep)
		usbInRegister(ep, regEPINT).Set(flags &^ epintTXFE)
		if flags&epintTXFE != 0 {
			handleUSBInFIFOEmpty(ep)
		}
		if flags&epintXFRC != 0 {
			handleUSBInTransfer(ep)
		}
	}
}

func usbInInterruptMask(ep uint32) uint32 {
	mask := usbRegister(regDIEPMSK).Get()
	if usbRegister(regDIEPEMPMSK).HasBits(1 << ep) {
		mask |= epintTXFE
	}
	return mask
}

func handleUSBInTransfer(ep uint32) {
	if ep == usb.CDC_ENDPOINT_IN {
		completeUSBEP3Transfer()
		return
	}
	if ep != 0 {
		if int(ep) < len(usbTxHandler) && usbTxHandler[ep] != nil {
			usbTxHandler[ep]()
		}
		return
	}
	if stm32USBEP0Offset < len(stm32USBEP0Data) {
		startUSBEP0Packet()
		return
	}
	stm32USBEP0Data = nil
	stm32USBEP0Offset = 0
	if stm32USBStatusIN {
		stm32USBStatusIN = false
		applyUSBAddress()
		armEP0Setup()
		return
	}
	stm32USBStatusOUT = true
	armUSBControlOUT()
}

func handleUSBInFIFOEmpty(ep uint32) {
	if ep != usb.CDC_ENDPOINT_IN || !stm32USBEP3Transfer.active {
		usbRegister(regDIEPEMPMSK).ClearBits(1 << ep)
		return
	}
	remaining := len(stm32USBEP3Transfer.data) - stm32USBEP3Transfer.offset
	words := usbInRegister(ep, regDTXFSTS).Get() & dtxfstsWordsMask
	count := usbFIFOChunkSize(remaining, words)
	start := stm32USBEP3Transfer.offset
	writeUSBFIFO(ep, stm32USBEP3Transfer.data[start:start+count])
	stm32USBEP3Transfer.offset += count
	if stm32USBEP3Transfer.offset == len(stm32USBEP3Transfer.data) {
		usbRegister(regDIEPEMPMSK).ClearBits(1 << ep)
	}
}

func completeUSBEP3Transfer() {
	usbRegister(regDIEPEMPMSK).ClearBits(1 << usb.CDC_ENDPOINT_IN)
	if !stm32USBEP3Transfer.active {
		return
	}
	if stm32USBEP3Transfer.zlp {
		stm32USBEP3Transfer.zlp = false
		startUSBEP3Hardware(0, 1)
		return
	}
	stm32USBEP3Transfer = stm32USBInTransfer{}
	if usbTxHandler[usb.CDC_ENDPOINT_IN] != nil {
		usbTxHandler[usb.CDC_ENDPOINT_IN]()
	}
}

func handleUSBOutComplete() {
	pending := (usbRegister(regDAINT).Get() & usbRegister(regDAINTMSK).Get()) >> 16
	for ep := uint32(0); ep < usbEndpointCount; ep++ {
		if pending&(1<<ep) == 0 {
			continue
		}
		flags := usbOutRegister(ep, regEPINT).Get() & usbRegister(regDOEPMSK).Get()
		usbOutRegister(ep, regEPINT).Set(flags)
		if ep == 0 && flags&epintSTUP != 0 && stm32USBSetupReady {
			stm32USBSetupReady = false
			handleUSBSetup()
		}
		if ep == 0 && flags&epintXFRC != 0 && stm32USBStatusOUT {
			stm32USBStatusOUT = false
			armEP0Setup()
		}
		if ep != 0 && flags&epintXFRC != 0 {
			deliverUSBOUT(ep)
		}
	}
}

func deliverUSBOUT(ep uint32) {
	count := int(stm32USBOutCount[ep])
	buffer := udd_ep_out_cache_buffer[ep][:count]
	stm32USBOutCount[ep] = 0
	if usbRxHandler[ep] == nil || usbRxHandler[ep](buffer) {
		AckUsbOutTransfer(ep)
	}
}

func initEndpoint(ep, config uint32) {
	if ep >= usbEndpointCount || config == usb.ENDPOINT_TYPE_DISABLE {
		return
	}
	if ep == 0 {
		initUSBControlEndpoint()
		return
	}
	typeBits := (config & 3) << epctlEPTYPPos
	if config&usb.EndpointIn != 0 {
		value := usbPacketSize | typeBits | ep<<epctlTXFNUMPos | epctlSD0PID | epctlUSBAEP
		usbInRegister(ep, regEPCTL).Set(value)
		usbRegister(regDAINTMSK).SetBits(1 << ep)
		if ep == usb.CDC_ENDPOINT_IN {
			retireAbortedUSBEP3Transfer()
		}
		return
	}
	value := usbPacketSize | typeBits | epctlSD0PID | epctlUSBAEP
	usbOutRegister(ep, regEPCTL).Set(value)
	usbRegister(regDAINTMSK).SetBits(1 << (16 + ep))
	armUSBOUT(ep)
}

func initUSBControlEndpoint() {
	usbInRegister(0, regEPCTL).Set(epctlUSBAEP)
	usbOutRegister(0, regEPCTL).Set(epctlUSBAEP)
	usbRegister(regDAINTMSK).SetBits(1 | 1<<16)
}

func retireAbortedUSBEP3Transfer() {
	if !stm32USBEP3Aborted || usbTxHandler[usb.CDC_ENDPOINT_IN] == nil {
		return
	}
	stm32USBEP3Aborted = false
	usbTxHandler[usb.CDC_ENDPOINT_IN]()
}

func armEP0Setup() {
	usbOutRegister(0, regEPTSIZ).Set(eptsizSTUPCNT3 | eptsizPKTCNT1 | 24)
	usbOutRegister(0, regEPCTL).SetBits(epctlCNAK | epctlEPENA | epctlUSBAEP)
}

func armUSBControlOUT() {
	usbOutRegister(0, regEPTSIZ).Set(eptsizPKTCNT1 | usbPacketSize)
	usbOutRegister(0, regEPCTL).SetBits(epctlCNAK | epctlEPENA | epctlUSBAEP)
}

func armUSBOUT(ep uint32) {
	usbOutRegister(ep, regEPTSIZ).Set(eptsizPKTCNT1 | usbPacketSize)
	usbOutRegister(ep, regEPCTL).SetBits(epctlCNAK | epctlEPENA)
}

func handleUSBSetAddress(setup usb.Setup) bool {
	stm32USBAddress = setup.WValueL & 0x7f
	stm32USBAddressPending = true
	usbRegister(regDCFG).ReplaceBits(uint32(stm32USBAddress)<<4, dcfgDADMask, 0)
	SendZlp()
	return true
}

func applyUSBAddress() {
	if !stm32USBAddressPending {
		return
	}
	usbRegister(regDCFG).ReplaceBits(uint32(stm32USBAddress)<<4, dcfgDADMask, 0)
	stm32USBAddressPending = false
}

// SendUSBInPacket starts a control or CDC bulk-IN transfer.
func SendUSBInPacket(ep uint32, data []byte) bool {
	if ep == 0 {
		sendUSBPacket(ep, data)
		return true
	}
	if ep == usb.CDC_ENDPOINT_IN {
		return startUSBEP3Transfer(data)
	}
	if int(ep) < len(usbTxHandler) && usbTxHandler[ep] != nil {
		usbTxHandler[ep]()
	}
	return true
}

func startUSBEP3Transfer(data []byte) bool {
	size, packets, zlp, ok := planUSBINTransfer(len(data))
	if !ok || stm32USBEP3Transfer.active {
		return false
	}
	stm32USBEP3Transfer = stm32USBInTransfer{data: data, zlp: zlp, active: true}
	startUSBEP3Hardware(size, packets)
	return true
}

func startUSBEP3Hardware(size, packets uint32) {
	ep := uint32(usb.CDC_ENDPOINT_IN)
	value := size&eptsizXFRMask | packets<<19&eptsizPKTCNTMask
	usbInRegister(ep, regEPTSIZ).Set(value)
	usbInRegister(ep, regEPCTL).SetBits(epctlCNAK | epctlEPENA)
	if size > 0 {
		usbRegister(regDIEPEMPMSK).SetBits(1 << ep)
	}
}

func planUSBINTransfer(length int) (uint32, uint32, bool, bool) {
	if length < 0 || length > int(usbPacketSize*usbINPacketCountMax) {
		return 0, 0, false, false
	}
	if length == 0 {
		return 0, 1, false, true
	}
	size := uint32(length)
	packets := (size + usbPacketSize - 1) / usbPacketSize
	return size, packets, size%usbPacketSize == 0, true
}

func usbFIFOChunkSize(remaining int, freeWords uint32) int {
	count := remaining
	if count > int(usbPacketSize) {
		count = int(usbPacketSize)
	}
	if capacity := int(freeWords * 4); count > capacity {
		count = capacity
	}
	return count
}

//go:noinline
func sendUSBPacket(ep uint32, data []byte) {
	if ep != 0 {
		return
	}
	stm32USBEP0Data = data
	stm32USBEP0Offset = 0
	startUSBEP0Packet()
}

func startUSBEP0Packet() {
	remaining := len(stm32USBEP0Data) - stm32USBEP0Offset
	count := remaining
	if count > int(usbPacketSize) {
		count = int(usbPacketSize)
	}
	packet := stm32USBEP0Data[stm32USBEP0Offset : stm32USBEP0Offset+count]
	stm32USBEP0Offset += count
	usbInRegister(0, regEPTSIZ).Set(eptsizPKTCNT1 | uint32(count)&eptsizXFRMask)
	usbInRegister(0, regEPCTL).SetBits(epctlCNAK | epctlEPENA | epctlUSBAEP)
	writeUSBFIFO(0, packet)
}

func writeUSBFIFO(ep uint32, data []byte) {
	for offset := 0; offset < len(data); offset += 4 {
		var word uint32
		for index := 0; index < 4 && offset+index < len(data); index++ {
			word |= uint32(data[offset+index]) << (8 * uint(index))
		}
		usbTransmitFIFO(ep).Set(word)
	}
}

func ReceiveUSBControlPacket() ([cdcLineInfoSize]byte, error) {
	var result [cdcLineInfoSize]byte
	armUSBControlOUT()
	defer armEP0Setup()
	for count := uint32(0); count < usbControlPollLimit; count++ {
		if usbRegister(regGINTSTS).Get()&gintRXFLVL == 0 {
			continue
		}
		status := decodeSTM32USBReceiveStatus(usbRegister(regGRXSTSP).Get())
		if status.packet == rxPacketData && status.ep == 0 {
			readUSBRXFIFO(result[:], status.count)
			if status.count != cdcLineInfoSize {
				return result, ErrUSBBytesRead
			}
			return result, nil
		}
		handleUSBReceiveStatus(status)
	}
	return result, ErrUSBReadTimeout
}

func handleEndpointRx(ep uint32) []byte {
	count := int(stm32USBOutCount[ep])
	return udd_ep_out_cache_buffer[ep][:count]
}

// AckUsbOutTransfer prepares an OUT endpoint for another packet.
func AckUsbOutTransfer(ep uint32) {
	ep &= 0x7f
	if ep == 0 {
		armEP0Setup()
		return
	}
	armUSBOUT(ep)
}

func SendZlp() {
	sendUSBPacket(0, nil)
}

// SetStallEPIn stalls an IN endpoint.
func (dev *USBDevice) SetStallEPIn(ep uint32) {
	ep &= 0x7f
	if ep < usbEndpointCount {
		usbInRegister(ep, regEPCTL).SetBits(epctlSTALL)
	}
}

// SetStallEPOut stalls an OUT endpoint.
func (dev *USBDevice) SetStallEPOut(ep uint32) {
	ep &= 0x7f
	if ep < usbEndpointCount {
		usbOutRegister(ep, regEPCTL).SetBits(epctlSTALL)
	}
}

// ClearStallEPIn clears an IN endpoint stall.
func (dev *USBDevice) ClearStallEPIn(ep uint32) {
	ep &= 0x7f
	if ep < usbEndpointCount {
		usbInRegister(ep, regEPCTL).ClearBits(epctlSTALL)
		usbInRegister(ep, regEPCTL).SetBits(epctlSD0PID | epctlCNAK)
	}
}

// ClearStallEPOut clears an OUT endpoint stall.
func (dev *USBDevice) ClearStallEPOut(ep uint32) {
	ep &= 0x7f
	if ep < usbEndpointCount {
		usbOutRegister(ep, regEPCTL).ClearBits(epctlSTALL)
		usbOutRegister(ep, regEPCTL).SetBits(epctlSD0PID | epctlCNAK)
	}
}
