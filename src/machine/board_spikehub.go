//go:build spikehub

package machine

import "device/stm32"

const (
	PA5  = portA + 5
	PA6  = portA + 6
	PA7  = portA + 7
	PA15 = portA + 15
	PB15 = portB + 15

	SPI0_SCK_PIN = PA5
	SPI0_SDI_PIN = PA6
	SPI0_SDO_PIN = PA7
	LAT          = PA15
	GSCLK        = PB15
	AF5_SPI1     = 5
	AF9_TIM12    = 9
)

var (
	SPI1 = &SPI{Bus: stm32.SPI1, AltFuncSelector: AF5_SPI1}
	SPI0 = SPI1
)

// 8 MHz at HSI16 is below the TLC5955 33 MHz limit (pybricks@76bfc071: platform.c; TLC5955).
const GSCLKFrequency = 8_000_000

func ConfigureGSCLK() {
	GSCLK.ConfigureAltFunc(PinConfig{Mode: PinModePWMOutput}, AF9_TIM12)
	stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM12EN)
	stm32.RCC.APB1RSTR.SetBits(stm32.RCC_APB1RSTR_TIM12RST)
	stm32.RCC.APB1RSTR.ClearBits(stm32.RCC_APB1RSTR_TIM12RST)
	top := uint32(APB1_TIM_FREQ / GSCLKFrequency)
	stm32.TIM12.PSC.Set(0)
	stm32.TIM12.ARR.Set(top - 1)
	stm32.TIM12.CCR2.Set(top / 2)
	stm32.TIM12.CCMR1_Output.Set(6 << 12)
	stm32.TIM12.CCER.Set(1 << 4)
	stm32.TIM12.EGR.Set(stm32.TIM_EGR_UG)
	stm32.TIM12.CR1.Set(stm32.TIM_CR1_ARPE | stm32.TIM_CR1_CEN)
}
