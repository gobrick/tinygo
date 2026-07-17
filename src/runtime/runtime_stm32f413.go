//go:build stm32f413

package runtime

import (
	"device/arm"
	"device/stm32"
	"machine"
)

const (
	applicationFlashOrigin = 0x08010000
	HCLK_FREQ_HZ           = 16000000
	PCLK1_FREQ_HZ          = HCLK_FREQ_HZ
	PCLK2_FREQ_HZ          = HCLK_FREQ_HZ
)

func init() {
	bootloaderHandoffAssumptions()
	initVectorTable()
	initClockHSI16()
	initTickTimer(&machine.TIM3)
}

// Bootloader handoff assumptions:
// - MSP is loaded from this image's vector word 0.
// - No peripheral IRQ remains enabled or pending.
// - RCC state may be arbitrary and is replaced below.
// - Cortex-M startup copies .data and clears .bss.
func bootloaderHandoffAssumptions() {}

func initVectorTable() {
	arm.SCB.VTOR.Set(applicationFlashOrigin)
}

func initClockHSI16() {
	stm32.RCC.CR.ReplaceBits(16, stm32.RCC_CR_HSITRIM_Msk, stm32.RCC_CR_HSITRIM_Pos)
	stm32.RCC.CR.SetBits(stm32.RCC_CR_HSION)
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_HSIRDY) {
	}

	stm32.RCC.CFGR.Set(0)
	for stm32.RCC.CFGR.Get()&stm32.RCC_CFGR_SWS_Msk != 0 {
	}

	stm32.FLASH.ACR.Set(stm32.FLASH_ACR_ICEN | stm32.FLASH_ACR_DCEN | stm32.FLASH_ACR_PRFTEN)
	for stm32.FLASH.ACR.Get()&stm32.FLASH_ACR_LATENCY_Msk != 0 {
	}

	stm32.RCC.CR.ClearBits(stm32.RCC_CR_HSEON | stm32.RCC_CR_CSSON | stm32.RCC_CR_PLLON)
	stm32.RCC.CIR.Set(0)
}

func putchar(byte) {}

func getchar() byte { return 0 }

func buffered() int { return 0 }
