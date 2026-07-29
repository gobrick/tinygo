//go:build cortexm && spikehub && !nxp && !qemu

package runtime

import (
	"device/arm"
	"machine"
)

func exit(code int) {
	machine.StopPorts()
	halt()
}

func abort() {
	machine.StopPorts()
	halt()
}

func halt() {
	for {
		arm.Asm("wfi")
	}
}
