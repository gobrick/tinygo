//go:build stm32f413

package machine

import (
	"device/stm32"
	"errors"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

var deviceIDAddr = []uintptr{0x1FFF7A10, 0x1FFF7A14, 0x1FFF7A18}

const (
	PA0  = portA + 0
	PA1  = portA + 1
	PA2  = portA + 2
	PA3  = portA + 3
	PA4  = portA + 4
	PA5  = portA + 5
	PA6  = portA + 6
	PA7  = portA + 7
	PA8  = portA + 8
	PA9  = portA + 9
	PA10 = portA + 10
	PA11 = portA + 11
	PA12 = portA + 12
	PA13 = portA + 13
	PA14 = portA + 14
	PA15 = portA + 15

	PB0  = portB + 0
	PB1  = portB + 1
	PB2  = portB + 2
	PB3  = portB + 3
	PB4  = portB + 4
	PB5  = portB + 5
	PB6  = portB + 6
	PB7  = portB + 7
	PB8  = portB + 8
	PB9  = portB + 9
	PB10 = portB + 10
	PB11 = portB + 11
	PB12 = portB + 12
	PB13 = portB + 13
	PB14 = portB + 14
	PB15 = portB + 15

	PC0  = portC + 0
	PC1  = portC + 1
	PC2  = portC + 2
	PC3  = portC + 3
	PC4  = portC + 4
	PC5  = portC + 5
	PC6  = portC + 6
	PC7  = portC + 7
	PC8  = portC + 8
	PC9  = portC + 9
	PC10 = portC + 10
	PC11 = portC + 11
	PC12 = portC + 12
	PC13 = portC + 13
	PC14 = portC + 14
	PC15 = portC + 15

	PD0  = portD + 0
	PD1  = portD + 1
	PD2  = portD + 2
	PD3  = portD + 3
	PD4  = portD + 4
	PD5  = portD + 5
	PD6  = portD + 6
	PD7  = portD + 7
	PD8  = portD + 8
	PD9  = portD + 9
	PD10 = portD + 10
	PD11 = portD + 11
	PD12 = portD + 12
	PD13 = portD + 13
	PD14 = portD + 14
	PD15 = portD + 15

	PE0  = portE + 0
	PE1  = portE + 1
	PE2  = portE + 2
	PE3  = portE + 3
	PE4  = portE + 4
	PE5  = portE + 5
	PE6  = portE + 6
	PE7  = portE + 7
	PE8  = portE + 8
	PE9  = portE + 9
	PE10 = portE + 10
	PE11 = portE + 11
	PE12 = portE + 12
	PE13 = portE + 13
	PE14 = portE + 14
	PE15 = portE + 15
)

func CPUFrequency() uint32 {
	return 16000000
}

const APB1_TIM_FREQ = 16000000
const APB2_TIM_FREQ = 16000000

func (p Pin) getPort() *stm32.GPIO_Type {
	switch p / 16 {
	case 0:
		return stm32.GPIOA
	case 1:
		return stm32.GPIOB
	case 2:
		return stm32.GPIOC
	case 3:
		return stm32.GPIOD
	case 4:
		return stm32.GPIOE
	case 5:
		return stm32.GPIOF
	case 6:
		return stm32.GPIOG
	case 7:
		return stm32.GPIOH
	default:
		panic("machine: unknown port")
	}
}

func (p Pin) enableClock() {
	port := p / 16
	if port > 7 {
		panic("machine: unknown port")
	}
	stm32.RCC.AHB1ENR.SetBits(1 << port)
}

var TIM2 = TIM{
	EnableRegister: &stm32.RCC.APB1ENR,
	EnableFlag:     stm32.RCC_APB1ENR_TIM2EN,
	Device:         stm32.TIM2,
	busFreq:        APB1_TIM_FREQ,
}

func (t *TIM) registerUPInterrupt() interrupt.Interrupt {
	if t == &TIM2 {
		return interrupt.New(stm32.IRQ_TIM2, TIM2.handleUPInterrupt)
	}
	return interrupt.Interrupt{}
}

func (t *TIM) registerOCInterrupt() interrupt.Interrupt {
	if t == &TIM2 {
		return interrupt.New(stm32.IRQ_TIM2, TIM2.handleOCInterrupt)
	}
	return interrupt.Interrupt{}
}

func (t *TIM) enableMainOutput() {}

type arrtype = uint32
type psctype = uint32
type arrRegType = volatile.Register32

const ARR_MAX = 0x10000
const PSC_MAX = 0x10000

func initRNG() {
	stm32.RCC.AHB2ENR.SetBits(stm32.RCC_AHB2ENR_RNGEN)
	stm32.RNG.CR.SetBits(stm32.RNG_CR_RNGEN)
}

type SPI struct {
	Bus             *stm32.SPI_Type
	AltFuncSelector uint8
}

func (spi *SPI) config8Bits() {
	spi.Bus.CR1.ClearBits(stm32.SPI_CR1_DFF)
}

// Reset pulse: the bootloader may leave CR1/CR2/DMA state behind.
func (spi *SPI) resetPeripheral() {
	if unsafe.Pointer(spi.Bus) == unsafe.Pointer(stm32.SPI1) {
		stm32.RCC.APB2RSTR.SetBits(1 << stm32.RCC_APB2RSTR_SPI1RST_Pos)
		stm32.RCC.APB2RSTR.ClearBits(1 << stm32.RCC_APB2RSTR_SPI1RST_Pos)
		spi.Bus.CR1.Set(0)
		spi.Bus.CR2.Set(0)
	}
}

func (spi *SPI) configurePins(config SPIConfig) {
	spi.resetPeripheral()
	config.SCK.ConfigureAltFunc(PinConfig{Mode: PinModeSPICLK}, spi.AltFuncSelector)
	config.SDO.ConfigureAltFunc(PinConfig{Mode: PinModeSPISDO}, spi.AltFuncSelector)
	config.SDI.ConfigureAltFunc(PinConfig{Mode: PinModeSPISDI}, spi.AltFuncSelector)
	config.SCK.configurePushPullNoPullHighSpeed()
	config.SDO.configurePushPullNoPullHighSpeed()
}

func (p Pin) configurePushPullNoPullHighSpeed() {
	pin := uint8(p % 16)
	pos := pin * 2
	port := p.getPort()
	port.OTYPER.ReplaceBits(stm32.GPIO_OTYPER_OT0_PushPull, uint32(1), pin)
	port.PUPDR.ReplaceBits(gpioPullFloating, gpioPullMask, pos)
	port.OSPEEDR.ReplaceBits(gpioOutputSpeedHigh, gpioOutputSpeedMask, pos)
}

func (spi *SPI) getBaudRate(config SPIConfig) uint32 {
	clock := uint32(16000000)
	freq := config.Frequency
	if freq < clock/256 {
		freq = clock / 256
	}
	if freq > clock/2 {
		freq = clock / 2
	}
	divisor := uint32(2)
	br := uint32(0)
	for divisor < 256 && clock/divisor > freq {
		divisor *= 2
		br++
	}
	return br << stm32.SPI_CR1_BR_Pos
}

// Tx sends w and reads into r. A write-only transfer (r == nil) uses a
// transmit-only path that waits on TXE/BSY (not RXNE) and returns an error if
// TXE stalls, so the caller can skip latching an incomplete frame.
func (spi *SPI) Tx(w, r []byte) error {
	switch {
	case r == nil:
		return spi.transmit(w)
	case w == nil:
		for i := range r {
			b, err := spi.Transfer(0)
			if err != nil {
				return err
			}
			r[i] = b
		}
		return nil
	default:
		if len(w) != len(r) {
			return ErrTxInvalidSliceSize
		}
		for i, b := range w {
			got, err := spi.Transfer(b)
			if err != nil {
				return err
			}
			r[i] = got
		}
		return nil
	}
}

// spiWaitLimit bounds each SPI status spin: a few milliseconds at 16MHz, far
// above a byte time (~1us) yet imperceptible in the display loop.
const spiWaitLimit = 1 << 12

var errSPITimeout = errors.New("machine: SPI transmit timeout")

// transmit clocks out every byte transmit-only. A per-byte TXE stall aborts with
// an error so the caller skips the latch; the terminal TXE/BSY can stay asserted
// on this write-only path once the frame has shifted, so a timeout there is fine.
func (spi *SPI) transmit(w []byte) error {
	dr := (*volatile.Register8)(unsafe.Pointer(&spi.Bus.DR.Reg))
	for _, b := range w {
		if !spi.awaitSR(stm32.SPI_SR_TXE, true) {
			return errSPITimeout
		}
		dr.Set(b)
	}
	spi.awaitSR(stm32.SPI_SR_TXE, true)
	spi.awaitSR(stm32.SPI_SR_BSY, false)
	_ = spi.Bus.DR.Get()
	_ = spi.Bus.SR.Get()
	return nil
}

// awaitSR spins until SR flag equals set (returns true) or gives up after
// spiWaitLimit iterations (returns false) so a stuck status bit cannot hang.
func (spi *SPI) awaitSR(flag uint32, set bool) bool {
	for i := 0; i < spiWaitLimit; i++ {
		if spi.Bus.SR.HasBits(flag) == set {
			return true
		}
	}
	return false
}

func enableAltFuncClock(bus unsafe.Pointer) {
	if bus == unsafe.Pointer(stm32.SPI1) {
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_SPI1EN)
	}
}
