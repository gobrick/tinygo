//go:build stm32f413 && !spikehubcore

package machine

// EnterBootloader is the CDC handler's name for the 1200-baud touch. On this
// board it asks for update mode rather than a ROM loader, and the reset waits
// for the control transfer in progress to finish.
func EnterBootloader() { ScheduleSelfUpdateReset() }
