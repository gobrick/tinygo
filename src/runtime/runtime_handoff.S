.syntax unified
.cpu cortex-m4
.thumb
.eabi_attribute 7, 77

.section .text.handoff_menu_tail,"ax",%progbits
.balign 4

.global handoff_menu_tail
.type handoff_menu_tail,%function
.thumb_func
handoff_menu_tail:
    mrs r1, ipsr
    cmp r1, #0
    bne handoff_wrong_mode
    mov r11, r0
    cpsid i
    bl kinetigo_stop_ports

handoff_stop_tim6:
    movw r1, #0x1000
    movt r1, #0x4000
    movs r2, #0
    str r2, [r1]
    str r2, [r1, #0x0c]
    str r2, [r1, #0x10]
    movw r3, #0x3800
    movt r3, #0x4002
    ldr r2, [r3, #0x20]
    orr r2, r2, #0x10
    str r2, [r3, #0x20]
    bic r2, r2, #0x10
    str r2, [r3, #0x20]
    ldr r2, [r3, #0x40]
    bic r2, r2, #0x10
    str r2, [r3, #0x40]

handoff_stop_dac_dma:
    movw r1, #0x7400
    movt r1, #0x4000
    ldr r2, [r1, #0x34]
    ldr r2, [r1]
    bic r2, r2, #0x1000
    str r2, [r1]

handoff_stop_dma1_stream5:
    movw r0, #0
    movt r0, #0x20
    ldr r2, [r3, #0x30]
    orr r2, r2, r0
    str r2, [r3, #0x30]
    movw r1, #0x6088
    movt r1, #0x4002
    movs r2, #0
    str r2, [r1]
    dsb
    movw r1, #0x6000
    movt r1, #0x4002
    movw r2, #0x0f40
    str r2, [r1, #0x0c]
    ldr r2, [r3, #0x10]
    orr r2, r2, r0
    str r2, [r3, #0x10]
    bic r2, r2, r0
    str r2, [r3, #0x10]
    movw r2, #0x0f40
    str r2, [r1, #0x0c]
    ldr r2, [r3, #0x30]
    bic r2, r2, r0
    str r2, [r3, #0x30]

handoff_reset_dac:
    movw r0, #0
    movt r0, #0x2000
    ldr r2, [r3, #0x20]
    orr r2, r2, r0
    str r2, [r3, #0x20]
    bic r2, r2, r0
    str r2, [r3, #0x20]
    ldr r2, [r3, #0x40]
    bic r2, r2, r0
    str r2, [r3, #0x40]

handoff_stop_speaker_gpio:
    movw r1, #0
    movt r1, #0x4002
    movw r2, #0
    movt r2, #0x10
    str r2, [r1, #0x18]
    movw r3, #0x0300
    ldr r2, [r1]
    bic r2, r2, r3
    str r2, [r1]
    ldr r2, [r1, #0x08]
    bic r2, r2, r3
    str r2, [r1, #0x08]
    ldr r2, [r1, #0x0c]
    bic r2, r2, r3
    str r2, [r1, #0x0c]
    movs r3, #0x10
    ldr r2, [r1, #0x04]
    bic r2, r2, r3
    str r2, [r1, #0x04]
    movw r3, #0
    movt r3, #0x000f
    ldr r2, [r1, #0x20]
    bic r2, r2, r3
    str r2, [r1, #0x20]
    adds r1, #0x800
    movw r2, #0
    movt r2, #0x0400
    str r2, [r1, #0x18]
    movw r3, #0
    movt r3, #0x0030
    ldr r2, [r1]
    bic r2, r2, r3
    str r2, [r1]
    ldr r2, [r1, #0x08]
    bic r2, r2, r3
    str r2, [r1, #0x08]
    ldr r2, [r1, #0x0c]
    bic r2, r2, r3
    str r2, [r1, #0x0c]
    movw r3, #0x0400
    ldr r2, [r1, #0x04]
    bic r2, r2, r3
    str r2, [r1, #0x04]
    movw r3, #0x0f00
    ldr r2, [r1, #0x24]
    bic r2, r2, r3
    str r2, [r1, #0x24]

handoff_stop_tim7:
    movw r1, #0x1400
    movt r1, #0x4000
    movs r2, #0
    str r2, [r1]
    str r2, [r1, #0x0c]
    str r2, [r1, #0x10]
    movw r3, #0x3800
    movt r3, #0x4002
    ldr r2, [r3, #0x20]
    orr r2, r2, #0x20
    str r2, [r3, #0x20]
    bic r2, r2, #0x20
    str r2, [r3, #0x20]
    ldr r2, [r3, #0x40]
    bic r2, r2, #0x20
    str r2, [r3, #0x40]

handoff_stop_adc:
    movw r1, #0x2000
    movt r1, #0x4001
    movs r2, #0
    str r2, [r1, #0x08]
    str r2, [r1, #0x04]
    ldr r0, [r1, #0x4c]
    str r2, [r1]
    ldr r2, [r3, #0x24]
    orr r2, r2, #0x100
    str r2, [r3, #0x24]
    bic r2, r2, #0x100
    str r2, [r3, #0x24]
    ldr r2, [r3, #0x44]
    bic r2, r2, #0x100
    str r2, [r3, #0x44]

handoff_stop_button_gpio:
    movw r1, #0
    movt r1, #0x4002
    movw r2, #0
    movt r2, #2
    str r2, [r1, #0x18]
    movs r3, #0x0c
    ldr r2, [r1]
    bic r2, r2, r3
    str r2, [r1]
    ldr r2, [r1, #0x08]
    bic r2, r2, r3
    str r2, [r1, #0x08]
    ldr r2, [r1, #0x0c]
    bic r2, r2, r3
    str r2, [r1, #0x0c]
    movs r3, #2
    ldr r2, [r1, #0x04]
    bic r2, r2, r3
    str r2, [r1, #0x04]
    movs r3, #0xf0
    ldr r2, [r1, #0x20]
    bic r2, r2, r3
    str r2, [r1, #0x20]
    adds r1, #0x800
    movw r2, #0
    movt r2, #0x10
    str r2, [r1, #0x18]
    movw r3, #0x0300
    ldr r2, [r1]
    bic r2, r2, r3
    str r2, [r1]
    ldr r2, [r1, #0x08]
    bic r2, r2, r3
    str r2, [r1, #0x08]
    ldr r2, [r1, #0x0c]
    bic r2, r2, r3
    str r2, [r1, #0x0c]
    movs r3, #0x10
    ldr r2, [r1, #0x04]
    bic r2, r2, r3
    str r2, [r1, #0x04]
    movw r3, #0
    movt r3, #0x000f
    ldr r2, [r1, #0x20]
    bic r2, r2, r3
    str r2, [r1, #0x20]
    b handoff_after_menu_cleanup

.size handoff_menu_tail, .-handoff_menu_tail

.section .text.handoff_tail,"ax",%progbits
.balign 4

.global handoff_tail
.type handoff_tail,%function
.thumb_func
handoff_tail:
    // Admit hand-off only from Thread mode.
handoff_assert_thread:
    mrs r1, ipsr
    cmp r1, #0
    bne handoff_wrong_mode
    mov r11, r0

    // Keep configurable interrupts masked through the branch.
handoff_mask_interrupts:
    cpsid i
    bl kinetigo_stop_ports

handoff_after_menu_cleanup:
    // Stop SysTick before any pending state is cleared.
handoff_stop_systick:
    movw r1, #0xe010
    movt r1, #0xe000
    movs r2, #0
    str r2, [r1]

    // Disconnect and reset OTG FS.
handoff_reset_otg:
    movw r1, #0x0804
    movt r1, #0x5000
    ldr r2, [r1]
    orr r2, r2, #2
    str r2, [r1]
    movw r1, #0x0008
    movt r1, #0x5000
    ldr r2, [r1]
    bic r2, r2, #1
    str r2, [r1]
    movw r1, #0x0038
    movt r1, #0x5000
    ldr r2, [r1]
    bic r2, r2, #0x00010000
    str r2, [r1]
    movw r3, #0x3800
    movt r3, #0x4002
    ldr r2, [r3, #0x14]
    orr r2, r2, #0x80
    str r2, [r3, #0x14]
    bic r2, r2, #0x80
    str r2, [r3, #0x14]
    ldr r2, [r3, #0x34]
    bic r2, r2, #0x80
    str r2, [r3, #0x34]

    // Disable PLLI2S after OTG FS is reset.
handoff_disable_plli2s:
    ldr r2, [r3]
    movw r1, #0
    movt r1, #0x0400
    bic r2, r2, r1
    str r2, [r3]
    movt r1, #0x0800
    movw r0, #0xffff

    // Bound the wait for PLLI2S to report stopped.
handoff_wait_plli2s:
    ldr r2, [r3]
    tst r2, r1
    beq handoff_stop_tim12
    subs r0, #1
    bne handoff_wait_plli2s

    // Reset and stop the matrix clock timer.
handoff_stop_tim12:
    ldr r2, [r3, #0x20]
    orr r2, r2, #0x40
    str r2, [r3, #0x20]
    bic r2, r2, #0x40
    str r2, [r3, #0x20]
    ldr r2, [r3, #0x40]
    bic r2, r2, #0x40
    str r2, [r3, #0x40]

    // Reset and stop the display SPI.
handoff_stop_spi:
    ldr r2, [r3, #0x24]
    movw r1, #0x1000
    orr r2, r2, r1
    str r2, [r3, #0x24]
    bic r2, r2, r1
    str r2, [r3, #0x24]
    ldr r2, [r3, #0x44]
    bic r2, r2, r1
    str r2, [r3, #0x44]

    // Return the display GPIO pins to their reset configuration.
handoff_stop_gpio:
    movw r1, #0
    movt r1, #0x4002
    movw r2, #0
    movt r2, #0x8000
    str r2, [r1, #0x18]
    movw r3, #0xfc00
    movt r3, #0xc000
    ldr r2, [r1]
    bic r2, r2, r3
    str r2, [r1]
    ldr r2, [r1, #0x08]
    bic r2, r2, r3
    str r2, [r1, #0x08]
    ldr r2, [r1, #0x0c]
    bic r2, r2, r3
    str r2, [r1, #0x0c]
    movw r3, #0x80e0
    ldr r2, [r1, #0x04]
    bic r2, r2, r3
    str r2, [r1, #0x04]
    movw r3, #0
    movt r3, #0xfff0
    ldr r2, [r1, #0x20]
    bic r2, r2, r3
    str r2, [r1, #0x20]
    movt r3, #0xf000
    ldr r2, [r1, #0x24]
    bic r2, r2, r3
    str r2, [r1, #0x24]
    adds r1, #0x400
    ldr r2, [r1]
    bic r2, r2, r3
    str r2, [r1]
    ldr r2, [r1, #0x08]
    bic r2, r2, r3
    str r2, [r1, #0x08]
    ldr r2, [r1, #0x0c]
    bic r2, r2, r3
    str r2, [r1, #0x0c]
    movw r3, #0x8000
    ldr r2, [r1, #0x04]
    bic r2, r2, r3
    str r2, [r1, #0x04]
    movw r3, #0
    movt r3, #0xf000
    ldr r2, [r1, #0x24]
    bic r2, r2, r3
    str r2, [r1, #0x24]

    // Disable every configurable interrupt.
handoff_disable_nvic:
    movw r1, #0xe180
    movt r1, #0xe000
    movw r2, #0xffff
    movt r2, #0xffff
    str r2, [r1, #0x00]
    str r2, [r1, #0x04]
    str r2, [r1, #0x08]
    str r2, [r1, #0x0c]
    str r2, [r1, #0x10]
    str r2, [r1, #0x14]
    str r2, [r1, #0x18]
    str r2, [r1, #0x1c]

    // Clear every pending configurable interrupt.
handoff_clear_nvic:
    movw r1, #0xe280
    movt r1, #0xe000
    str r2, [r1, #0x00]
    str r2, [r1, #0x04]
    str r2, [r1, #0x08]
    str r2, [r1, #0x0c]
    str r2, [r1, #0x10]
    str r2, [r1, #0x14]
    str r2, [r1, #0x18]
    str r2, [r1, #0x1c]

    // Clear pending SysTick and PendSV.
handoff_clear_system_pending:
    movw r1, #0xed04
    movt r1, #0xe000
    movw r2, #0
    movt r2, #0x0a00
    str r2, [r1]

    // Clear priority masks while keeping PRIMASK set.
handoff_clear_masks:
    movs r2, #0
    msr basepri, r2
    msr faultmask, r2

    // Install the slot vectors and enter on its MSP.
handoff_enter_slot:
    ldr r1, [r11]
    ldr r2, [r11, #4]
    msr msp, r1
    movw r3, #0xed08
    movt r3, #0xe000
    str r11, [r3]
    dsb
    isb
    movs r3, #0
    msr control, r3
    isb
    msr psp, r3
    bx r2

    // Refuse a call made from Handler mode.
handoff_wrong_mode:
    b handoff_wrong_mode

.size handoff_tail, .-handoff_tail
