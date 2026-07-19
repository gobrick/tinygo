//go:build stm32f413

package runtime

import (
	"device/arm"
	"device/stm32"
	"machine"
	_ "machine/usb/cdc"
)

const (
	applicationFlashOrigin = 0x08008000
	HCLK_FREQ_HZ           = 16000000
	PCLK1_FREQ_HZ          = HCLK_FREQ_HZ
	PCLK2_FREQ_HZ          = HCLK_FREQ_HZ
)

func init() {
	arm.DisableInterrupts()
	bootloaderHandoffAssumptions()
	initVectorTable()
	normalizeExceptions()
	initClockHSI16()
	resetTickTimer()
	initTickTimer(&machine.TIM2)
	machine.TIM2.UpInterrupt.SetPriority(0xc0)
	machine.InitSerial()
	arm.EnableInterrupts(0)
}

// Bootloader handoff: MSP loads from vector word 0; a fault before this init
// still vectors through the loader table until VTOR is set below.
func bootloaderHandoffAssumptions() {}

func initVectorTable() {
	arm.SCB.VTOR.Set(applicationFlashOrigin)
	arm.Asm("dsb")
	arm.Asm("isb")
}

func normalizeExceptions() {
	arm.SYST.SYST_CSR.ClearBits(arm.SYST_CSR_TICKINT | arm.SYST_CSR_ENABLE)
	arm.SCB.ICSR.SetBits(arm.SCB_ICSR_PENDSTCLR | arm.SCB_ICSR_PENDSVCLR)
	for word := range arm.NVIC.ICER {
		arm.NVIC.ICER[word].Set(0xffffffff)
		arm.NVIC.ICPR[word].Set(0xffffffff)
	}
	arm.SCB.AIRCR.Set(0x5fa<<arm.SCB_AIRCR_VECTKEY_Pos | 3<<arm.SCB_AIRCR_PRIGROUP_Pos)
	arm.AsmFull("msr BASEPRI, {b}", map[string]interface{}{"b": uint32(0)})
}

func initClockHSI16() {
	stm32.RCC.CR.ReplaceBits(16, stm32.RCC_CR_HSITRIM_Msk, stm32.RCC_CR_HSITRIM_Pos)
	stm32.RCC.CR.SetBits(stm32.RCC_CR_HSION)
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_HSIRDY) {
	}

	stm32.RCC.CFGR.ReplaceBits(
		stm32.RCC_CFGR_PPRE1_Div4<<stm32.RCC_CFGR_PPRE1_Pos|
			stm32.RCC_CFGR_PPRE2_Div4<<stm32.RCC_CFGR_PPRE2_Pos,
		stm32.RCC_CFGR_PPRE1_Msk|stm32.RCC_CFGR_PPRE2_Msk|stm32.RCC_CFGR_SW_Msk, 0)
	for stm32.RCC.CFGR.Get()&stm32.RCC_CFGR_SWS_Msk != 0 {
	}
	stm32.RCC.CFGR.Set(0)

	stm32.FLASH.ACR.ClearBits(stm32.FLASH_ACR_LATENCY_Msk)
	stm32.FLASH.ACR.SetBits(stm32.FLASH_ACR_ICEN | stm32.FLASH_ACR_DCEN | stm32.FLASH_ACR_PRFTEN)

	stm32.RCC.CR.ClearBits(stm32.RCC_CR_HSEON | stm32.RCC_CR_CSSON | stm32.RCC_CR_PLLON)
	stm32.RCC.CIR.Set(stm32.RCC_CIR_LSIRDYC | stm32.RCC_CIR_LSERDYC | stm32.RCC_CIR_HSIRDYC |
		stm32.RCC_CIR_HSERDYC | stm32.RCC_CIR_PLLRDYC | stm32.RCC_CIR_CSSC)
	stm32.RCC.CIR.Set(0)
}

func resetTickTimer() {
	stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM2EN)
	stm32.RCC.APB1RSTR.SetBits(stm32.RCC_APB1RSTR_TIM2RST)
	stm32.RCC.APB1RSTR.ClearBits(stm32.RCC_APB1RSTR_TIM2RST)
}

func putchar(value byte) {
	_ = machine.Serial.WriteByte(value)
}

func getchar() byte { return 0 }

func buffered() int { return 0 }
