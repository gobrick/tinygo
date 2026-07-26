//go:build stm32f413

package machine

import (
	"device/arm"
	"device/stm32"
	_ "unsafe"
)

// SelfUpdateTouchBaud is the line rate that, together with a DTR high-to-low
// edge, asks a running program to reboot into update mode.
const SelfUpdateTouchBaud = 1200

// The request is a word and its complement, so indeterminate memory after a
// cold start cannot look like one.
const (
	selfUpdateToken           = uint32(0x4b475544)
	selfUpdateTokenComplement = ^uint32(0x4b475544)
	selfUpdateIdentity        = uint32(0x4b474944)
	selfUpdateIdentityInverse = ^uint32(0x4b474944)
	selfUpdateMaxImage        = uint32(256 << 10)
	selfUpdateFlushTicks      = int64(2_000_000_000 / 16)
)

// coldResetFlags are the reset causes after which nothing a running program
// stored can be trusted, whatever the backup registers happen to hold.
//
// The pin flag is deliberately not among them. Measured on the hub: a software
// reset reports 0x14000000, software and pin together, because the internal
// reset controller drives NRST low. Rejecting on the pin flag rejects every
// real request. A button press on its own is still refused, because that does
// not set the software flag this requires.
const coldResetFlags = stm32.RCC_CSR_PORRSTF | stm32.RCC_CSR_BORRSTF |
	stm32.RCC_CSR_WDGRSTF | stm32.RCC_CSR_WWDGRSTF | stm32.RCC_CSR_LPWRRSTF

var selfUpdateResetPending bool

// ScheduleSelfUpdateReset records a one-shot request and arms a reset for the
// end of the current control transfer. Resetting from the setup handler would
// abort the status stage the host is still waiting on.
func ScheduleSelfUpdateReset() {
	enableBackupDomain()
	stm32.RTC.SetBKP0R(selfUpdateToken)
	stm32.RTC.SetBKP1R(selfUpdateTokenComplement)
	if stm32.RTC.GetBKP0R() != selfUpdateToken || stm32.RTC.GetBKP1R() != selfUpdateTokenComplement {
		return
	}
	stm32.RCC.CSR.SetBits(stm32.RCC_CSR_RMVF)
	selfUpdateResetPending = true
}

// completeSelfUpdateReset resets once the status stage has actually gone out.
func completeSelfUpdateReset() {
	if !selfUpdateResetPending {
		return
	}
	selfUpdateResetPending = false
	arm.SystemReset()
}

// TakeSelfUpdateRequest reports a genuine request exactly once. It clears the
// request before returning, so a transfer that fails leaves the hub coming back
// as the old application instead of looping into update mode.
func TakeSelfUpdateRequest() bool {
	flags := stm32.RCC.CSR.Get()
	stm32.RCC.CSR.SetBits(stm32.RCC_CSR_RMVF)

	enableBackupDomain()
	token, complement := stm32.RTC.GetBKP0R(), stm32.RTC.GetBKP1R()
	stm32.RTC.SetBKP0R(0)
	stm32.RTC.SetBKP1R(0)

	if token != selfUpdateToken || complement != selfUpdateTokenComplement {
		return false
	}
	return flags&stm32.RCC_CSR_SFTRSTF != 0 && flags&coldResetFlags == 0
}

// SelfUpdateResetFlags reports the reset causes latched at startup. It exists
// so a probe can say what the hardware actually did rather than infer it.
func SelfUpdateResetFlags() uint32 {
	return stm32.RCC.CSR.Get()
}

// SelfUpdateBackupControl reports the backup domain control register, so a
// probe can show whether the domain came back configured or reset.
func SelfUpdateBackupControl() uint32 {
	enableBackupDomain()
	return stm32.RCC.BDCR.Get()
}

// SelfUpdateTokenWords reports the backup words as found, for the same reason.
func SelfUpdateTokenWords() (uint32, uint32) {
	enableBackupDomain()
	return stm32.RTC.GetBKP0R(), stm32.RTC.GetBKP1R()
}

// FeedSelfUpdateWatchdog reloads an already-running independent watchdog.
func FeedSelfUpdateWatchdog() {
	stm32.IWDG.KR.Set(iwdgKeyReset)
}

// ArmSelfUpdateIdentity records the image span that must identify itself after reset.
func ArmSelfUpdateIdentity(length uint32) bool {
	enableBackupDomain()
	stm32.RTC.SetBKP2R(0)
	stm32.RTC.SetBKP3R(selfUpdateIdentityInverse)
	stm32.RTC.SetBKP4R(length)
	stm32.RTC.SetBKP2R(selfUpdateIdentity)
	return stm32.RTC.GetBKP2R() == selfUpdateIdentity &&
		stm32.RTC.GetBKP3R() == selfUpdateIdentityInverse &&
		stm32.RTC.GetBKP4R() == length
}

// TakeSelfUpdateIdentity returns one pending post-reset image report.
func TakeSelfUpdateIdentity() (uint32, bool) {
	enableBackupDomain()
	token := stm32.RTC.GetBKP2R()
	inverse := stm32.RTC.GetBKP3R()
	length := stm32.RTC.GetBKP4R()
	stm32.RTC.SetBKP2R(0)
	stm32.RTC.SetBKP3R(0)
	stm32.RTC.SetBKP4R(0)
	validLength := length >= 4 && length <= selfUpdateMaxImage && length&3 == 0
	return length, token == selfUpdateIdentity &&
		inverse == selfUpdateIdentityInverse && validLength
}

// SelfUpdatePowerOK reports whether VDD is above the roughly 3.14 V rising PVD threshold.
func SelfUpdatePowerOK() bool {
	stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_PWREN)
	stm32.PWR.CR.ReplaceBits(7, stm32.PWR_CR_PLS_Msk>>stm32.PWR_CR_PLS_Pos, stm32.PWR_CR_PLS_Pos)
	stm32.PWR.CR.SetBits(stm32.PWR_CR_PVDE)
	return stm32.PWR.GetCSR_PVDO() == stm32.PWR_CSR_PVDO_Higher
}

// FlushCDCOutput pushes queued console output onto the wire. A reset discards
// whatever is still sitting in the transmit ring, so a result printed and then
// reset away is a result nobody sees.
func FlushCDCOutput() bool {
	output, ok := USBCDC.(interface {
		DTR() bool
		Flush() bool
	})
	if !ok {
		return false
	}
	deadline := selfUpdateTicks() + selfUpdateFlushTicks
	for {
		if output.Flush() {
			return true
		}
		if !output.DTR() || selfUpdateTicks() >= deadline {
			return false
		}
		FeedSelfUpdateWatchdog()
	}
}

// DiscardCDCInput drops anything the host sent before the session began, so a
// stale byte cannot be mistaken for the start of a frame.
func DiscardCDCInput() {
	for USBCDC.Buffered() > 0 {
		if _, err := USBCDC.ReadByte(); err != nil {
			return
		}
	}
}

// enableBackupDomain unlocks the registers that survive a system reset. Neither
// the clock nor the write protection survives one, so both are re-established
// on every pass rather than assumed to still be in place.
func enableBackupDomain() {
	stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_PWREN)
	stm32.PWR.CR.SetBits(stm32.PWR_CR_DBP)
}

//go:linkname selfUpdateTicks runtime.ticks
func selfUpdateTicks() int64
