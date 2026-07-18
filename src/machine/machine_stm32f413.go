//go:build stm32f413

package machine

import (
	"device/stm32"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

var deviceIDAddr = []uintptr{0x1FFF7A10, 0x1FFF7A14, 0x1FFF7A18}

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
	pos := uint8(p%16) * 2
	port := p.getPort()
	port.OTYPER.ReplaceBits(stm32.GPIO_OTYPER_OT0_PushPull, stm32.GPIO_OTYPER_OT0_Msk, pos/2)
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

func enableAltFuncClock(bus unsafe.Pointer) {
	if bus == unsafe.Pointer(stm32.SPI1) {
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_SPI1EN)
	}
}
