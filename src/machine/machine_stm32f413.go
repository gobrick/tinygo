//go:build stm32f413

package machine

import (
	"device/stm32"
	"runtime/interrupt"
	"runtime/volatile"
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

var TIM3 = TIM{
	EnableRegister: &stm32.RCC.APB1ENR,
	EnableFlag:     stm32.RCC_APB1ENR_TIM3EN,
	Device:         stm32.TIM3,
	busFreq:        APB1_TIM_FREQ,
}

func (t *TIM) registerUPInterrupt() interrupt.Interrupt {
	if t == &TIM3 {
		return interrupt.New(stm32.IRQ_TIM3, TIM3.handleUPInterrupt)
	}
	return interrupt.Interrupt{}
}

func (t *TIM) registerOCInterrupt() interrupt.Interrupt {
	if t == &TIM3 {
		return interrupt.New(stm32.IRQ_TIM3, TIM3.handleOCInterrupt)
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
