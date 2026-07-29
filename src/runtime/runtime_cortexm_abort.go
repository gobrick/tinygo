//go:build cortexm && !nxp && !qemu && !spikehub

package runtime

import (
	"device/arm"
)

func exit(code int) {
	abort()
}

func abort() {
	// lock up forever
	for {
		arm.Asm("wfi")
	}
}
