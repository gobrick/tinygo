//go:build spikehub

package machine

import "device/stm32"

const (
	SPI0_SCK_PIN = PA5
	SPI0_SDI_PIN = PA6
	SPI0_SDO_PIN = PA7
	LAT          = PA15
	GSCLK        = PB15
	AF5_SPI1     = 5
	AF9_TIM12    = 9
	AF10_OTG_FS  = 10

	USBCDC_DM_PIN = PA11
	USBCDC_DP_PIN = PA12

	HSEFrequency = 16_000_000

	usb_STRING_PRODUCT      = "KinetiGo Hub"
	usb_STRING_MANUFACTURER = "KinetiGo"
)

var (
	usb_VID uint16 = 0x1209
	usb_PID uint16 = 0x0001
)

var (
	SPI1 = &SPI{Bus: stm32.SPI1, AltFuncSelector: AF5_SPI1}
	SPI0 = SPI1
)

// 8 MHz at HSI16 is below the TLC5955 33 MHz limit (pybricks@76bfc071: platform.c; TLC5955).
const GSCLKFrequency = 8_000_000

const (
	centreButtonThreshold  = (4084 + 3148) / 2
	centreButtonHysteresis = 16
	boardDelayTimerHz      = 1_000_000
	boardDelayTimerPSC     = APB1_TIM_FREQ/boardDelayTimerHz - 1
	shutdownHighReload     = 35
	shutdownLowReload      = 47
	shutdownToneMicros     = 60_000
	shutdownDACMask        = 0xffff
	shutdownDACEnable      = 1 << 0
	shutdownDACTrigger     = 1 << 2
	shutdownDACTriangle    = 2 << 6
	shutdownDACAmplitude   = 7 << 8
	shutdownDACDMAEnable   = 1 << 12
	shutdownDACDMAUnderrun = 1 << 13
	shutdownDMAStream      = 5
	shutdownDMAEnable      = 1 << 0
	shutdownDMAFlags       = 0x0f40
)

// HoldPower keeps the hub powered; it powers down if the latch is not held.
func HoldPower() {
	stm32.RCC.AHB1ENR.SetBits(stm32.RCC_AHB1ENR_GPIOAEN)
	stm32.GPIOA.BSRR.Set(1 << uint8(PA13%16))
	stm32.GPIOA.MODER.ReplaceBits(gpioModeOutput, gpioModeMask, uint8(PA13%16)*2)
}

// CentreHeldAtStartup samples the measured centre-button levels after settling.
func CentreHeldAtStartup() bool {
	stm32.RCC.AHB1ENR.SetBits(stm32.RCC_AHB1ENR_GPIOCEN)
	stm32.GPIOC.MODER.ReplaceBits(3, 3, 8)
	stm32.GPIOC.PUPDR.ReplaceBits(0, 3, 8)
	stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_ADC1EN)
	stm32.RCC.APB2RSTR.SetBits(stm32.RCC_APB2RSTR_ADCRST)
	stm32.RCC.APB2RSTR.ClearBits(stm32.RCC_APB2RSTR_ADCRST)
	stm32.ADC1.CR1.Set(0)
	stm32.ADC1.CR2.Set(0)
	stm32.ADC1.SR.Set(0)
	stm32.ADC1.SQR1.Set(0)
	stm32.ADC1.SQR2.Set(0)
	stm32.ADC1.SQR3.Set(14)
	stm32.ADC1.SMPR1.ReplaceBits(adcSampleCycles, 7, 12)
	stm32.ADC1.CR2.Set(stm32.ADC_CR2_ADON)
	first := sampleCentreButton()
	startBoardDelayTimer()
	waitBoardMicros(30_000)
	second := sampleCentreButton()
	stopBoardDelayTimer()
	stm32.ADC1.CR2.Set(0)
	stm32.ADC1.SR.Set(0)
	stm32.RCC.APB2RSTR.SetBits(stm32.RCC_APB2RSTR_ADCRST)
	stm32.RCC.APB2RSTR.ClearBits(stm32.RCC_APB2RSTR_ADCRST)
	stm32.RCC.APB2ENR.ClearBits(stm32.RCC_APB2ENR_ADC1EN)
	stm32.GPIOC.MODER.ReplaceBits(0, 3, 8)
	stm32.GPIOC.PUPDR.ReplaceBits(0, 3, 8)
	stm32.RCC.AHB1ENR.ClearBits(stm32.RCC_AHB1ENR_GPIOCEN)
	pressed := uint16(centreButtonThreshold - centreButtonHysteresis)
	return first <= pressed && second <= pressed
}

// PowerOff drops the hub power latch.
func PowerOff() {
	stm32.GPIOA.BSRR.Set(1 << (uint8(PA13%16) + 16))
}

// Shutdown plays the descending cue fully, then drops the power latch.
func Shutdown() {
	StopPorts()
	stm32.RCC.AHB1ENR.SetBits(stm32.RCC_AHB1ENR_DMA1EN | stm32.RCC_AHB1ENR_GPIOAEN | stm32.RCC_AHB1ENR_GPIOCEN)
	stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_DACEN | stm32.RCC_APB1ENR_TIM6EN)
	stm32.GPIOC.BSRR.Set(1 << 26)
	stm32.GPIOC.MODER.ReplaceBits(1, 3, 20)
	stm32.GPIOC.OTYPER.ClearBits(1 << 10)
	stm32.GPIOC.PUPDR.ReplaceBits(0, 3, 20)
	stm32.GPIOA.MODER.ReplaceBits(3, 3, 8)
	stm32.GPIOA.PUPDR.ReplaceBits(0, 3, 8)
	stm32.TIM6.CR1.Set(0)
	stm32.TIM6.DIER.Set(0)
	stm32.TIM6.SR.Set(0)
	underrun := stopShutdownDMA()
	enableShutdownDAC(underrun)
	setShutdownDACBias()
	startBoardDelayTimer()
	prepareShutdownTone(shutdownHighReload)
	waitBoardMicros(2_000)
	stm32.GPIOC.BSRR.Set(1 << 10)
	stm32.TIM6.EGR.Set(stm32.TIM_EGR_UG)
	stm32.TIM6.CR1.Set(stm32.TIM_CR1_ARPE | stm32.TIM_CR1_CEN)
	waitBoardMicros(shutdownToneMicros)
	stm32.TIM6.ARR.Set(shutdownLowReload)
	waitBoardMicros(shutdownToneMicros)
	stm32.GPIOC.BSRR.Set(1 << 26)
	stm32.TIM6.CR1.Set(0)
	setShutdownDACBias()
	PowerOff()
}

func sampleCentreButton() uint16 {
	stm32.ADC1.CR2.SetBits(stm32.ADC_CR2_SWSTART)
	for !stm32.ADC1.SR.HasBits(stm32.ADC_SR_EOC) {
	}
	result := uint16(stm32.ADC1.DR.Get())
	stm32.ADC1.SR.Set(0)
	return result
}

func startBoardDelayTimer() {
	stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM5EN)
	stm32.RCC.APB1RSTR.SetBits(stm32.RCC_APB1RSTR_TIM5RST)
	stm32.RCC.APB1RSTR.ClearBits(stm32.RCC_APB1RSTR_TIM5RST)
	stm32.TIM5.CR1.Set(0)
	stm32.TIM5.PSC.Set(uint32(boardDelayTimerPSC))
	stm32.TIM5.ARR.Set(0xffffffff)
	stm32.TIM5.CNT.Set(0)
	stm32.TIM5.EGR.Set(stm32.TIM_EGR_UG)
	stm32.TIM5.SR.Set(0)
	stm32.TIM5.CR1.Set(stm32.TIM_CR1_CEN)
}

func stopBoardDelayTimer() {
	stm32.TIM5.CR1.Set(0)
	stm32.RCC.APB1RSTR.SetBits(stm32.RCC_APB1RSTR_TIM5RST)
	stm32.RCC.APB1RSTR.ClearBits(stm32.RCC_APB1RSTR_TIM5RST)
	stm32.RCC.APB1ENR.ClearBits(stm32.RCC_APB1ENR_TIM5EN)
}

func waitBoardMicros(duration uint32) {
	stm32.TIM5.CNT.Set(0)
	for stm32.TIM5.CNT.Get() < duration {
	}
}

func prepareShutdownTone(reload uint32) {
	stm32.TIM6.CR1.Set(0)
	stm32.TIM6.PSC.Set(0)
	stm32.TIM6.ARR.Set(reload)
	stm32.TIM6.CNT.Set(0)
	stm32.TIM6.CR2.Set(stm32.TIM_CR2_MMS_Update << stm32.TIM_CR2_MMS_Pos)
	setShutdownDACControl(0)
	stm32.DAC.DHR12R1.Set(1921)
	control := uint32(shutdownDACTrigger | shutdownDACTriangle | shutdownDACAmplitude)
	setShutdownDACControl(control)
	setShutdownDACControl(control | shutdownDACEnable)
}

func setShutdownDACBias() {
	setShutdownDACControl(0)
	stm32.DAC.DHR12R1.Set(2048)
	setShutdownDACControl(shutdownDACEnable)
}

func setShutdownDACControl(control uint32) {
	current := stm32.DAC.CR.Get()
	stm32.DAC.CR.Set(current&^shutdownDACMask | control&shutdownDACMask)
}

func stopShutdownDMA() bool {
	underrun := stm32.DAC.SR.HasBits(shutdownDACDMAUnderrun)
	setShutdownDACControl(stm32.DAC.CR.Get() &^ shutdownDACDMAEnable)
	stream := &stm32.DMA1.ST[shutdownDMAStream]
	stream.CR.ClearBits(shutdownDMAEnable)
	for stream.CR.HasBits(shutdownDMAEnable) {
	}
	stream.CR.Set(0)
	stm32.DMA1.HIFCR.Set(shutdownDMAFlags)
	stm32.RCC.APB1ENR.ClearBits(stm32.RCC_APB1ENR_DACEN)
	return underrun
}

func enableShutdownDAC(underrun bool) {
	stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_DACEN)
	if underrun {
		stm32.DAC.SR.Set(shutdownDACDMAUnderrun)
	}
}

func ConfigureLAT() {
	LAT.Configure(PinConfig{Mode: PinOutput})
	LAT.configurePushPullNoPullHighSpeed()
	LAT.Low()
}

func ConfigureGSCLK() {
	GSCLK.ConfigureAltFunc(PinConfig{Mode: PinModePWMOutput}, AF9_TIM12)
	GSCLK.configurePushPullNoPullHighSpeed()
	stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM12EN)
	stm32.RCC.APB1RSTR.SetBits(stm32.RCC_APB1RSTR_TIM12RST)
	stm32.RCC.APB1RSTR.ClearBits(stm32.RCC_APB1RSTR_TIM12RST)
	stm32.TIM12.CR1.Set(0)
	stm32.TIM12.DIER.Set(0)
	stm32.TIM12.CCER.Set(0)
	stm32.TIM12.SR.Set(0)
	top := uint32(APB1_TIM_FREQ / GSCLKFrequency)
	stm32.TIM12.PSC.Set(0)
	stm32.TIM12.ARR.Set(top - 1)
	stm32.TIM12.CCR2.Set(top / 2)
	stm32.TIM12.CCMR1_Output.Set(6 << 12)
	stm32.TIM12.CCER.Set(1 << 4)
	stm32.TIM12.EGR.Set(stm32.TIM_EGR_UG)
	stm32.TIM12.CR1.Set(stm32.TIM_CR1_ARPE | stm32.TIM_CR1_CEN)
}
