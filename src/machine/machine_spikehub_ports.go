//go:build spikehub

package machine

import (
	"device/stm32"
	_ "unsafe"
)

const (
	AF2_TIM4      = 2
	AF8_UART5     = 8
	PCLK1_FREQ_HZ = uint32(16_000_000)
	PortPWMTop    = uint32(1333)
)

// PB8 and PB9 share one mode register, so both bridge inputs change together.
// A single-pin transition would leave one driving while the other is not,
// which is the maximum-drive condition.
const (
	portDModeMask    = uint32(0xf) << 16
	portDModeOutput  = uint32(0x5) << 16
	portDModeAlt     = uint32(0xa) << 16
	portDModeForward = uint32(0x6) << 16
	portDModeReverse = uint32(0x9) << 16
	portDSetBoth     = uint32(0x300)
	portDResetBoth   = uint32(0x300) << 16
)

// PortPinSet describes one external port's board wiring.
type PortPinSet struct {
	Motor1 Pin
	Motor2 Pin
	TX     Pin
	RX     Pin
	Enable Pin
	P5     Pin
	P6     Pin
}

// ExternalPortPins lists ports A through F.
var ExternalPortPins = [6]PortPinSet{
	{PE9, PE11, PE8, PE7, PA10, PD7, PD8},
	{PE13, PE14, PD1, PD0, PA8, PD9, PD10},
	{PB6, PB7, PE1, PE0, PE5, PD11, PE4},
	{PB8, PB9, PC12, PD2, PB2, PC15, PC14},
	{PC6, PC7, PE3, PE2, PB5, PC13, PE12},
	{PC8, PB1, PD15, PD14, PC5, PC11, PE6},
}

// PortReadStatus describes one nonblocking UART receive attempt.
type PortReadStatus uint8

const (
	PortNoData PortReadStatus = iota
	PortData
	PortFault
)

var (
	portDSession bool
	portDPWM     bool
	portDDriving bool
)

//go:linkname stopPorts kinetigo_stop_ports
func stopPorts()

// InitPorts establishes coast before enabling the shared port supply.
func InitPorts() {
	StopPorts()
	PB8.Low()
	PB9.Low()
	PB2.High()
	PA14.Low()
	PB8.Configure(PinConfig{Mode: PinOutput})
	PB9.Configure(PinConfig{Mode: PinOutput})
	PB2.Configure(PinConfig{Mode: PinOutput})
	PA14.Configure(PinConfig{Mode: PinOutput})
	PA14.High()
}

// StopPorts returns every external port to its safe electrical state.
func StopPorts() {
	stopPorts()
	portDSession = false
	portDPWM = false
	portDDriving = false
}

// PortDOpen connects port D to UART5 at the requested baud.
func PortDOpen(baud uint32) bool {
	StopPorts()
	InitPorts()
	if !PortDSetBaud(baud) {
		StopPorts()
		return false
	}
	PC12.ConfigureAltFunc(PinConfig{Mode: PinModeUARTTX}, AF8_UART5)
	PD2.ConfigureAltFunc(PinConfig{Mode: PinModeUARTRX}, AF8_UART5)
	PB2.Low()
	portDSession = true
	return true
}

// PortDSetBaud configures UART5 from the full APB1 peripheral clock.
func PortDSetBaud(baud uint32) bool {
	if baud == 0 {
		return false
	}
	divisor := (PCLK1_FREQ_HZ + baud/2) / baud
	if divisor < 16 || divisor > 0xffff {
		return false
	}
	PB2.High()
	stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_UART5EN)
	stm32.RCC.APB1RSTR.SetBits(stm32.RCC_APB1RSTR_UART5RST)
	stm32.RCC.APB1RSTR.ClearBits(stm32.RCC_APB1RSTR_UART5RST)
	stm32.UART5.CR1.Set(0)
	stm32.UART5.CR2.Set(0)
	stm32.UART5.CR3.Set(0)
	stm32.UART5.BRR.Set(divisor)
	clearPortDUART()
	stm32.UART5.CR1.Set(stm32.USART_CR1_TE | stm32.USART_CR1_RE | stm32.USART_CR1_UE)
	if portDSession {
		PB2.Low()
	}
	return true
}

// PortDProbeRX powers the port and samples the receive line as a plain input,
// reporting how many transitions it saw and the level it finished at.
func PortDProbeRX(samples int) (uint32, bool) {
	StopPorts()
	InitPorts()
	PB2.Low()
	PD2.Configure(PinConfig{Mode: PinInput})
	last := PD2.Get()
	changes := uint32(0)
	for i := 0; i < samples; i++ {
		if level := PD2.Get(); level != last {
			changes++
			last = level
		}
	}
	return changes, last
}

// PortDListen opens the port at one baud and counts how many bytes arrive
// cleanly against how many framing or noise errors occur.
func PortDListen(baud uint32, samples int) (uint32, uint32) {
	if !PortDOpen(baud) {
		return 0, 0
	}
	good, bad := uint32(0), uint32(0)
	for i := 0; i < samples; i++ {
		switch _, status := PortDRead(); status {
		case PortData:
			good++
		case PortFault:
			bad++
		}
	}
	return good, bad
}

// PortDRead performs one nonblocking receive attempt.
func PortDRead() (byte, PortReadStatus) {
	status := stm32.UART5.SR.Get()
	if status&0x0f != 0 {
		_ = stm32.UART5.DR.Get()
		return 0, PortFault
	}
	if status&stm32.USART_SR_RXNE == 0 {
		return 0, PortNoData
	}
	return byte(stm32.UART5.DR.Get()), PortData
}

// PortDWrite writes one byte only when the transmit register is empty.
func PortDWrite(b byte) bool {
	if !stm32.UART5.SR.HasBits(stm32.USART_SR_TXE) {
		return false
	}
	stm32.UART5.DR.Set(uint32(b))
	return true
}

// PortDTransmitComplete reports whether the shift register is empty.
func PortDTransmitComplete() bool {
	return stm32.UART5.SR.HasBits(stm32.USART_SR_TC)
}

// PortDDrive chops one bridge input and holds the other high, the way the
// reference driver does. Emulating the held pin with a compare of zero looked
// equivalent and was not.
func PortDDrive(duty uint32, forward bool) bool {
	if duty > PortPWMTop {
		return false
	}
	setupPortDPWM()
	if forward {
		stm32.TIM4.CCR3.Set(duty)
	} else {
		stm32.TIM4.CCR4.Set(duty)
	}
	stm32.TIM4.EGR.Set(stm32.TIM_EGR_UG)
	mode, hold := portDModeReverse, uint32(1)<<8
	if forward {
		mode, hold = portDModeForward, uint32(1)<<9
	}
	stm32.GPIOB.BSRR.Set(hold)
	stm32.GPIOB.AFRH.ReplaceBits(AF2_TIM4|AF2_TIM4<<4, 0xff, 0)
	stm32.GPIOB.OTYPER.ClearBits(portDSetBoth)
	stm32.GPIOB.PUPDR.ClearBits(portDModeMask)
	stm32.GPIOB.MODER.ReplaceBits(mode, portDModeMask, 0)
	portDDriving = true
	return true
}

// PortDCoast drives both bridge inputs low as GPIO outputs.
func PortDCoast() {
	setPortDBoth(portDResetBoth)
}

// PortDBrake drives both bridge inputs high as GPIO outputs.
func PortDBrake() {
	setPortDBoth(portDSetBoth)
}

func setPortDBoth(bsrr uint32) {
	stm32.GPIOB.BSRR.Set(bsrr)
	stm32.GPIOB.OTYPER.ClearBits(portDSetBoth)
	stm32.GPIOB.PUPDR.ClearBits(portDModeMask)
	stm32.GPIOB.MODER.ReplaceBits(portDModeOutput, portDModeMask, 0)
	portDDriving = false
}

func setupPortDPWM() {
	if portDPWM {
		return
	}
	PortDCoast()
	stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM4EN)
	stm32.RCC.APB1RSTR.SetBits(stm32.RCC_APB1RSTR_TIM4RST)
	stm32.RCC.APB1RSTR.ClearBits(stm32.RCC_APB1RSTR_TIM4RST)
	stm32.TIM4.CR1.Set(0)
	stm32.TIM4.DIER.Set(0)
	stm32.TIM4.CCER.Set(0)
	stm32.TIM4.PSC.Set(0)
	stm32.TIM4.ARR.Set(PortPWMTop - 1)
	stm32.TIM4.CCR3.Set(PortPWMTop)
	stm32.TIM4.CCR4.Set(PortPWMTop)
	mode := uint32(stm32.TIM_CCMR2_Output_OC3PE | stm32.TIM_CCMR2_Output_OC4PE)
	mode |= stm32.TIM_CCMR2_Output_OC3M_PwmMode1 << stm32.TIM_CCMR2_Output_OC3M_Pos
	mode |= stm32.TIM_CCMR2_Output_OC4M_PwmMode1 << stm32.TIM_CCMR2_Output_OC4M_Pos
	stm32.TIM4.CCMR2_Output.Set(mode)
	output := uint32(stm32.TIM_CCER_CC3E | stm32.TIM_CCER_CC3P)
	output |= stm32.TIM_CCER_CC4E | stm32.TIM_CCER_CC4P
	stm32.TIM4.CCER.Set(output)
	stm32.TIM4.EGR.Set(stm32.TIM_EGR_UG)
	stm32.TIM4.CR1.Set(stm32.TIM_CR1_ARPE | stm32.TIM_CR1_CEN)
	portDPWM = true
}

func clearPortDUART() {
	for stm32.UART5.SR.HasBits(stm32.USART_SR_RXNE) {
		_ = stm32.UART5.DR.Get()
	}
	_ = stm32.UART5.SR.Get()
	_ = stm32.UART5.DR.Get()
}
