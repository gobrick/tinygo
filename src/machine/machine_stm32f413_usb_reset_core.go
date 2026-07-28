//go:build stm32f413 && spikehubcore

package machine

// EnterBootloader restarts the core with an update request.
func EnterBootloader() { ScheduleSelfUpdateReset() }
