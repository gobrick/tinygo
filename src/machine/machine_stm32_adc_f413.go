//go:build stm32f413

package machine

import (
	"device/stm32"
	"runtime/interrupt"
)

const adcSampleCycles = uint32(stm32.ADC_SMPR2_SMP1_Cycles15)

// InitADC enables ADC1 in single-conversion, 12-bit mode.
func InitADC() {
	state := interrupt.Disable()
	stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_ADC1EN)
	stm32.RCC.APB2RSTR.SetBits(stm32.RCC_APB2RSTR_ADCRST)
	stm32.RCC.APB2RSTR.ClearBits(stm32.RCC_APB2RSTR_ADCRST)
	stm32.ADC1.CR1.Set(0)
	stm32.ADC1.CR2.Set(0)
	stm32.ADC1.SR.Set(0)
	stm32.ADC1.SQR1.Set(0)
	stm32.ADC1.SQR2.Set(0)
	stm32.ADC1.SQR3.Set(0)
	stm32.ADC1.CR2.Set(stm32.ADC_CR2_ADON)
	interrupt.Restore(state)
}

// Configure places an ADC pin in analogue mode with the verified sample time.
func (a ADC) Configure(ADCConfig) {
	state := interrupt.Disable()
	a.Pin.Configure(PinConfig{Mode: PinInputAnalog})
	channel := uint32(a.getChannel())
	if channel > 9 {
		stm32.ADC1.SMPR1.ReplaceBits(adcSampleCycles, 7, uint8((channel-10)*3))
	} else {
		stm32.ADC1.SMPR2.ReplaceBits(adcSampleCycles, 7, uint8(channel*3))
	}
	interrupt.Restore(state)
}

// Get performs one software-triggered conversion and returns the 12-bit result
// scaled to the machine ADC range.
func (a ADC) Get() uint16 {
	state := interrupt.Disable()
	stm32.ADC1.SQR3.Set(uint32(a.getChannel()))
	stm32.ADC1.CR2.SetBits(stm32.ADC_CR2_SWSTART)
	for !stm32.ADC1.SR.HasBits(stm32.ADC_SR_EOC) {
	}
	result := uint16(stm32.ADC1.DR.Get()) << 4
	stm32.ADC1.SR.Set(0)
	interrupt.Restore(state)
	return result
}

func (a ADC) getChannel() uint8 {
	switch a.Pin {
	case PA0:
		return 0
	case PA1:
		return 1
	case PA2:
		return 2
	case PA3:
		return 3
	case PA4:
		return 4
	case PA5:
		return 5
	case PA6:
		return 6
	case PA7:
		return 7
	case PB0:
		return 8
	case PB1:
		return 9
	case PC0:
		return 10
	case PC1:
		return 11
	case PC2:
		return 12
	case PC3:
		return 13
	case PC4:
		return 14
	case PC5:
		return 15
	default:
		return 0
	}
}
