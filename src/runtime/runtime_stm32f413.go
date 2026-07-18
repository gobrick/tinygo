//go:build stm32f413

package runtime

import (
	"device/arm"
	"device/stm32"
	"machine"
)

const (
	applicationFlashOrigin = 0x08008000
	HCLK_FREQ_HZ           = 16000000
	PCLK1_FREQ_HZ          = HCLK_FREQ_HZ
	PCLK2_FREQ_HZ          = HCLK_FREQ_HZ
)

func init() {
	bootloaderHandoffAssumptions()
	initVectorTable()
	initClockHSI16()
	resetTickTimer()
	initTickTimer(&machine.TIM2)
}

// Bootloader handoff assumptions:
// - MSP is loaded from this image's vector word 0.
// - No peripheral IRQ remains enabled or pending.
// - RCC state may be arbitrary and is replaced below.
// - Cortex-M startup copies .data and clears .bss.
// Residual window: a fault raised before this init runs vectors through
// the bootloader table; unavoidable without upstream Reset_Handler
// changes. Revisited with hardware evidence in PR 2.
func bootloaderHandoffAssumptions() {}

func initVectorTable() {
	arm.SCB.VTOR.Set(applicationFlashOrigin)
}

func initClockHSI16() {
	stm32.RCC.CR.ReplaceBits(16, stm32.RCC_CR_HSITRIM_Msk, stm32.RCC_CR_HSITRIM_Pos)
	stm32.RCC.CR.SetBits(stm32.RCC_CR_HSION)
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_HSIRDY) {
	}

	// Conservative APB divisors first: a hot handoff clock stays legal
	// while the switch to HSI drains.
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

func putchar(byte) {}

func getchar() byte { return 0 }

func buffered() int { return 0 }
