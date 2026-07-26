.syntax unified
.cpu cortex-m4
.thumb
.eabi_attribute 7, 77

.section .ramfuncs.commit,"ax",%progbits
.balign 4

.global commit_engine
.global commit_fail_reset

.type commit_engine,%function
.thumb_func
commit_engine:
    // Mask configurable interrupts and stop SysTick.
commit_mask_interrupts:
    cpsid i
    movw r8, #0xe010
    movt r8, #0xe000
    movw r9, #0
    movt r9, #0
    str.w r9, [r8]

    // Disable every configurable interrupt.
commit_disable_interrupts:
    movw r8, #0xe180
    movt r8, #0xe000
    movw r9, #0xffff
    movt r9, #0xffff
    str.w r9, [r8, #0x00]
    str.w r9, [r8, #0x04]
    str.w r9, [r8, #0x08]
    str.w r9, [r8, #0x0c]
    str.w r9, [r8, #0x10]
    str.w r9, [r8, #0x14]
    str.w r9, [r8, #0x18]
    str.w r9, [r8, #0x1c]

    // Clear every pending configurable interrupt.
commit_clear_interrupts:
    movw r8, #0xe280
    movt r8, #0xe000
    str.w r9, [r8, #0x00]
    str.w r9, [r8, #0x04]
    str.w r9, [r8, #0x08]
    str.w r9, [r8, #0x0c]
    str.w r9, [r8, #0x10]
    str.w r9, [r8, #0x14]
    str.w r9, [r8, #0x18]
    str.w r9, [r8, #0x1c]

    // Install and synchronize the SRAM vector table.
commit_install_vectors:
    movw r8, #0xed08
    movt r8, #0xe000
    movw r9, #0x1000
    movt r9, #0x2000
    str.w r9, [r8]
    dsb
    isb

    // Disable and reset both flash caches.
commit_reset_caches:
    movw r8, #0x3c00
    movt r8, #0x4002
    movw r9, #0
    movt r9, #0
    str.w r9, [r8]
    movw r8, #0x3c00
    movt r8, #0x4002
    movw r9, #0x1800
    movt r9, #0
    str.w r9, [r8]
    movw r8, #0x3c00
    movt r8, #0x4002
    movw r9, #0
    movt r9, #0
    str.w r9, [r8]

    // Wait for a prior flash operation while feeding the watchdog.
commit_initial_busy:
    movw r10, #0x3c0c
    movt r10, #0x4002
    movw r11, #0x3000
    movt r11, #0x4000
    movw r12, #0xaaaa
    movt r12, #0
    movw r5, #0
    movt r5, #1
commit_initial_wait:
    ldr.w r6, [r10]
    str.w r12, [r11]
    tst r6, r5
    bne.w commit_initial_wait

    // Unlock the flash controller with the fixed key sequence.
commit_unlock_flash:
    movw r8, #0x3c04
    movt r8, #0x4002
    movw r9, #0x0123
    movt r9, #0x4567
    str.w r9, [r8]
    movw r8, #0x3c04
    movt r8, #0x4002
    movw r9, #0x89ab
    movt r9, #0xcdef
    str.w r9, [r8]
    movw r8, #0x3c0c
    movt r8, #0x4002
    ldr.w r6, [r8]

    // Prove the admitted source length is inside the update bounds.
commit_bound_length:
    movw r7, #4
    movt r7, #0
    cmp r1, r7
    blo.w commit_failure_reset
    movw r7, #0
    movt r7, #4
    cmp r1, r7
    bhi.w commit_failure_reset

    // Prove the admitted source span is inside SRAM.
commit_bound_source:
    movw r7, #0
    movt r7, #0x2000
    cmp r0, r7
    blo.w commit_failure_reset
    movw r7, #0
    movt r7, #0x2005
    subs r7, r7, r1
    cmp r0, r7
    bhi.w commit_failure_reset

    // Set the fixed application base and erase-span bound.
commit_set_bounds:
    movw r3, #0x8000
    movt r3, #0x0800
    movw r7, #0x8000
    movt r7, #5

    // Erase sector 2 after supply and status admission.
commit_erase_sequence:
commit_erase_sector_2:
    movw r8, #0x7004
    movt r8, #0x4000
    ldr.w r6, [r8]
    movw r5, #4
    movt r5, #0
    tst r6, r5
    bne.w commit_failure_reset
    movw r8, #0x3c0c
    movt r8, #0x4002
    movw r9, #0x01f3
    movt r9, #0
    str.w r9, [r8]
    movw r8, #0x3c10
    movt r8, #0x4002
    movw r9, #0x0012
    movt r9, #1
    str.w r9, [r8]

    // Wait for sector 2 while feeding the watchdog.
commit_erase_sector_2_busy:
    movw r10, #0x3c0c
    movt r10, #0x4002
    movw r11, #0x3000
    movt r11, #0x4000
    movw r12, #0xaaaa
    movt r12, #0
    movw r5, #0
    movt r5, #1
commit_erase_sector_2_wait:
    ldr.w r6, [r10]
    str.w r12, [r11]
    tst r6, r5
    bne.w commit_erase_sector_2_wait

    // Clear sector 2 control bits and reject every error flag.
commit_erase_sector_2_status:
    movw r8, #0x3c10
    movt r8, #0x4002
    movw r9, #0
    movt r9, #0
    str.w r9, [r8]
    movw r8, #0x3c0c
    movt r8, #0x4002
    ldr.w r6, [r8]
    movw r5, #0x01f2
    movt r5, #0
    tst r6, r5
    bne.w commit_failure_reset

    // Erase sector 3 after supply and status admission.
commit_erase_sector_3:
    movw r8, #0x7004
    movt r8, #0x4000
    ldr.w r6, [r8]
    movw r5, #4
    movt r5, #0
    tst r6, r5
    bne.w commit_failure_reset
    movw r8, #0x3c0c
    movt r8, #0x4002
    movw r9, #0x01f3
    movt r9, #0
    str.w r9, [r8]
    movw r8, #0x3c10
    movt r8, #0x4002
    movw r9, #0x001a
    movt r9, #1
    str.w r9, [r8]

    // Wait for sector 3 while feeding the watchdog.
commit_erase_sector_3_busy:
    movw r10, #0x3c0c
    movt r10, #0x4002
    movw r11, #0x3000
    movt r11, #0x4000
    movw r12, #0xaaaa
    movt r12, #0
    movw r5, #0
    movt r5, #1
commit_erase_sector_3_wait:
    ldr.w r6, [r10]
    str.w r12, [r11]
    tst r6, r5
    bne.w commit_erase_sector_3_wait

    // Clear sector 3 control bits and reject every error flag.
commit_erase_sector_3_status:
    movw r8, #0x3c10
    movt r8, #0x4002
    movw r9, #0
    movt r9, #0
    str.w r9, [r8]
    movw r8, #0x3c0c
    movt r8, #0x4002
    ldr.w r6, [r8]
    movw r5, #0x01f2
    movt r5, #0
    tst r6, r5
    bne.w commit_failure_reset

    // Erase sector 4 after supply and status admission.
commit_erase_sector_4:
    movw r8, #0x7004
    movt r8, #0x4000
    ldr.w r6, [r8]
    movw r5, #4
    movt r5, #0
    tst r6, r5
    bne.w commit_failure_reset
    movw r8, #0x3c0c
    movt r8, #0x4002
    movw r9, #0x01f3
    movt r9, #0
    str.w r9, [r8]
    movw r8, #0x3c10
    movt r8, #0x4002
    movw r9, #0x0022
    movt r9, #1
    str.w r9, [r8]

    // Wait for sector 4 while feeding the watchdog.
commit_erase_sector_4_busy:
    movw r10, #0x3c0c
    movt r10, #0x4002
    movw r11, #0x3000
    movt r11, #0x4000
    movw r12, #0xaaaa
    movt r12, #0
    movw r5, #0
    movt r5, #1
commit_erase_sector_4_wait:
    ldr.w r6, [r10]
    str.w r12, [r11]
    tst r6, r5
    bne.w commit_erase_sector_4_wait

    // Clear sector 4 control bits and reject every error flag.
commit_erase_sector_4_status:
    movw r8, #0x3c10
    movt r8, #0x4002
    movw r9, #0
    movt r9, #0
    str.w r9, [r8]
    movw r8, #0x3c0c
    movt r8, #0x4002
    ldr.w r6, [r8]
    movw r5, #0x01f2
    movt r5, #0
    tst r6, r5
    bne.w commit_failure_reset

    // Erase sector 5 after supply and status admission.
commit_erase_sector_5:
    movw r8, #0x7004
    movt r8, #0x4000
    ldr.w r6, [r8]
    movw r5, #4
    movt r5, #0
    tst r6, r5
    bne.w commit_failure_reset
    movw r8, #0x3c0c
    movt r8, #0x4002
    movw r9, #0x01f3
    movt r9, #0
    str.w r9, [r8]
    movw r8, #0x3c10
    movt r8, #0x4002
    movw r9, #0x002a
    movt r9, #1
    str.w r9, [r8]

    // Wait for sector 5 while feeding the watchdog.
commit_erase_sector_5_busy:
    movw r10, #0x3c0c
    movt r10, #0x4002
    movw r11, #0x3000
    movt r11, #0x4000
    movw r12, #0xaaaa
    movt r12, #0
    movw r5, #0
    movt r5, #1
commit_erase_sector_5_wait:
    ldr.w r6, [r10]
    str.w r12, [r11]
    tst r6, r5
    bne.w commit_erase_sector_5_wait

    // Clear sector 5 control bits and reject every error flag.
commit_erase_sector_5_status:
    movw r8, #0x3c10
    movt r8, #0x4002
    movw r9, #0
    movt r9, #0
    str.w r9, [r8]
    movw r8, #0x3c0c
    movt r8, #0x4002
    ldr.w r6, [r8]
    movw r5, #0x01f2
    movt r5, #0
    tst r6, r5
    bne.w commit_failure_reset

    // Erase sector 6 after supply and status admission.
commit_erase_sector_6:
    movw r8, #0x7004
    movt r8, #0x4000
    ldr.w r6, [r8]
    movw r5, #4
    movt r5, #0
    tst r6, r5
    bne.w commit_failure_reset
    movw r8, #0x3c0c
    movt r8, #0x4002
    movw r9, #0x01f3
    movt r9, #0
    str.w r9, [r8]
    movw r8, #0x3c10
    movt r8, #0x4002
    movw r9, #0x0032
    movt r9, #1
    str.w r9, [r8]

    // Wait for sector 6 while feeding the watchdog.
commit_erase_sector_6_busy:
    movw r10, #0x3c0c
    movt r10, #0x4002
    movw r11, #0x3000
    movt r11, #0x4000
    movw r12, #0xaaaa
    movt r12, #0
    movw r5, #0
    movt r5, #1
commit_erase_sector_6_wait:
    ldr.w r6, [r10]
    str.w r12, [r11]
    tst r6, r5
    bne.w commit_erase_sector_6_wait

    // Clear sector 6 control bits and reject every error flag.
commit_erase_sector_6_status:
    movw r8, #0x3c10
    movt r8, #0x4002
    movw r9, #0
    movt r9, #0
    str.w r9, [r8]
    movw r8, #0x3c0c
    movt r8, #0x4002
    ldr.w r6, [r8]
    movw r5, #0x01f2
    movt r5, #0
    tst r6, r5
    bne.w commit_failure_reset

    // Initialize the full-span blank verification.
commit_blank_initialize:
    movw r2, #0
    movt r2, #0

    // Verify every erased byte while feeding the watchdog.
commit_blank_loop:
    cmp r2, r7
    bhs.w commit_after_blank
    movw r8, #0x3000
    movt r8, #0x4000
    movw r9, #0xaaaa
    movt r9, #0
    str.w r9, [r8]
    ldrb r4, [r3, r2]
    cmp r4, #0xff
    bne.w commit_failure_reset
    adds r2, #1
    b.w commit_blank_loop

    // Initialize byte programming after the blank span passes.
commit_after_blank:
    movw r2, #0
    movt r2, #0

    // Skip activation and erased source bytes in the body loop.
commit_program_loop:
    cmp r2, r1
    bhs.w commit_after_program
    movw r8, #0x3000
    movt r8, #0x4000
    movw r9, #0xaaaa
    movt r9, #0
    str.w r9, [r8]
    cmp r2, #3
    beq.w commit_program_increment
    ldrb r4, [r0, r2]
    cmp r4, #0xff
    beq.w commit_program_increment

    // Admit and arm one x8 body-byte program.
commit_program_byte:
    movw r8, #0x7004
    movt r8, #0x4000
    ldr.w r6, [r8]
    movw r5, #4
    movt r5, #0
    tst r6, r5
    bne.w commit_failure_reset
    movw r8, #0x3c0c
    movt r8, #0x4002
    movw r9, #0x01f3
    movt r9, #0
    str.w r9, [r8]
    movw r8, #0x3c10
    movt r8, #0x4002
    movw r9, #1
    movt r9, #0
    str.w r9, [r8]
    strb r4, [r3, r2]

    // Wait for one body byte while feeding the watchdog.
commit_program_byte_busy:
    movw r10, #0x3c0c
    movt r10, #0x4002
    movw r11, #0x3000
    movt r11, #0x4000
    movw r12, #0xaaaa
    movt r12, #0
    movw r5, #0
    movt r5, #1
commit_program_byte_wait:
    ldr.w r6, [r10]
    str.w r12, [r11]
    tst r6, r5
    bne.w commit_program_byte_wait

    // Clear body-program control bits and reject every error flag.
commit_program_byte_status:
    movw r8, #0x3c10
    movt r8, #0x4002
    movw r9, #0
    movt r9, #0
    str.w r9, [r8]
    movw r8, #0x3c0c
    movt r8, #0x4002
    ldr.w r6, [r8]
    movw r5, #0x01f2
    movt r5, #0
    tst r6, r5
    bne.w commit_failure_reset

    // Advance the body-program induction variable once.
commit_program_increment:
    adds r2, #1
    b.w commit_program_loop

    // Initialize the pre-activation image and tail comparison.
commit_after_program:
    movw r2, #0
    movt r2, #0

    // Compare the body and erased tail before activation.
commit_preverify_loop:
    cmp r2, r7
    bhs.w commit_activate
    movw r8, #0x3000
    movt r8, #0x4000
    movw r9, #0xaaaa
    movt r9, #0
    str.w r9, [r8]
    ldrb r5, [r3, r2]
    cmp r2, r1
    bhs.w commit_preverify_tail
    cmp r2, #3
    beq.w commit_preverify_tail
    ldrb r4, [r0, r2]
    cmp r5, r4
    bne.w commit_failure_reset
    b.w commit_preverify_increment

    // Require erased flash at activation and beyond the image.
commit_preverify_tail:
    cmp r5, #0xff
    bne.w commit_failure_reset

    // Advance the pre-activation comparison once.
commit_preverify_increment:
    adds r2, #1
    b.w commit_preverify_loop

    // Install the fault-to-poison vector table before activation.
commit_activate:
    movw r8, #0xed08
    movt r8, #0xe000
    movw r9, #0x1200
    movt r9, #0x2000
    str.w r9, [r8]
    dsb
    isb

    // Admit and program the activation byte last.
commit_activate_write:
    movw r8, #0x7004
    movt r8, #0x4000
    ldr.w r6, [r8]
    movw r5, #4
    movt r5, #0
    tst r6, r5
    bne.w commit_failure_reset
    movw r8, #0x3c0c
    movt r8, #0x4002
    movw r9, #0x01f3
    movt r9, #0
    str.w r9, [r8]
    movw r8, #0x3c10
    movt r8, #0x4002
    movw r9, #1
    movt r9, #0
    str.w r9, [r8]
    movw r4, #0x20
    movt r4, #0
    strb r4, [r3, #3]

    // Wait for activation while feeding the watchdog.
commit_activate_busy:
    movw r10, #0x3c0c
    movt r10, #0x4002
    movw r11, #0x3000
    movt r11, #0x4000
    movw r12, #0xaaaa
    movt r12, #0
    movw r5, #0
    movt r5, #1
commit_activate_wait:
    ldr.w r6, [r10]
    str.w r12, [r11]
    tst r6, r5
    bne.w commit_activate_wait

    // Clear activation control bits and poison on any error flag.
commit_activate_status:
    movw r8, #0x3c10
    movt r8, #0x4002
    movw r9, #0
    movt r9, #0
    str.w r9, [r8]
    movw r8, #0x3c0c
    movt r8, #0x4002
    ldr.w r6, [r8]
    movw r5, #0x01f2
    movt r5, #0
    tst r6, r5
    bne.w commit_poison_activation

    // Initialize the post-activation image and tail comparison.
commit_postverify_initialize:
    movw r2, #0
    movt r2, #0

    // Compare the complete image and erased tail after activation.
commit_postverify_loop:
    cmp r2, r7
    bhs.w commit_success_reset
    movw r8, #0x3000
    movt r8, #0x4000
    movw r9, #0xaaaa
    movt r9, #0
    str.w r9, [r8]
    ldrb r5, [r3, r2]
    cmp r2, r1
    bhs.w commit_postverify_tail
    ldrb r4, [r0, r2]
    cmp r5, r4
    bne.w commit_poison_activation
    b.w commit_postverify_increment

    // Require erased flash beyond the activated image.
commit_postverify_tail:
    cmp r5, #0xff
    bne.w commit_poison_activation

    // Advance the post-activation comparison once.
commit_postverify_increment:
    adds r2, #1
    b.w commit_postverify_loop

    // Lock the controller and request reset after full verification.
commit_success_reset:
    movw r8, #0x3c10
    movt r8, #0x4002
    movw r9, #0
    movt r9, #0x8000
    str.w r9, [r8]
    movw r8, #0xed0c
    movt r8, #0xe000
    movw r9, #4
    movt r9, #0x05fa
    str.w r9, [r8]
    dsb

    // Remain in SRAM if the successful reset is delayed.
commit_success_reset_terminal:
    b.w commit_success_reset_terminal

    // Wait for any interrupted operation before poisoning.
commit_poison_activation:
    movw r10, #0x3c0c
    movt r10, #0x4002
    movw r11, #0x3000
    movt r11, #0x4000
    movw r12, #0xaaaa
    movt r12, #0
    movw r5, #0
    movt r5, #1
commit_poison_entry_wait:
    ldr.w r6, [r10]
    str.w r12, [r11]
    tst r6, r5
    bne.w commit_poison_entry_wait
    movw r3, #0x8000
    movt r3, #0x0800

    // Feed and retry until poisoning can start on acceptable supply.
commit_poison_supply:
    movw r8, #0x3000
    movt r8, #0x4000
    movw r9, #0xaaaa
    movt r9, #0
    str.w r9, [r8]
    movw r8, #0x7004
    movt r8, #0x4000
    ldr.w r6, [r8]
    movw r5, #4
    movt r5, #0
    tst r6, r5
    bne.w commit_poison_activation

    // Program zero over the activation byte.
commit_poison_write:
    movw r8, #0x3c0c
    movt r8, #0x4002
    movw r9, #0x01f3
    movt r9, #0
    str.w r9, [r8]
    movw r8, #0x3c10
    movt r8, #0x4002
    movw r9, #1
    movt r9, #0
    str.w r9, [r8]
    movw r4, #0
    movt r4, #0
    strb r4, [r3, #3]

    // Wait for poisoning while feeding the watchdog.
commit_poison_busy:
    movw r10, #0x3c0c
    movt r10, #0x4002
    movw r11, #0x3000
    movt r11, #0x4000
    movw r12, #0xaaaa
    movt r12, #0
    movw r5, #0
    movt r5, #1
commit_poison_wait:
    ldr.w r6, [r10]
    str.w r12, [r11]
    tst r6, r5
    bne.w commit_poison_wait

    // Clear poison control bits and retry on any error flag.
commit_poison_status:
    movw r8, #0x3c10
    movt r8, #0x4002
    movw r9, #0
    movt r9, #0
    str.w r9, [r8]
    movw r8, #0x3c0c
    movt r8, #0x4002
    ldr.w r6, [r8]
    movw r5, #0x01f2
    movt r5, #0
    tst r6, r5
    bne.w commit_poison_activation

    // Reread until activation is no longer 0x20.
commit_poison_verify:
    ldrb r4, [r3, #3]
    cmp r4, #0x20
    beq.w commit_poison_activation
    b.w commit_failure_reset

    // Lock the controller and request reset on a safe failure path.
commit_failure_reset:
    movw r8, #0x3c10
    movt r8, #0x4002
    movw r9, #0
    movt r9, #0x8000
    str.w r9, [r8]
    movw r8, #0xed0c
    movt r8, #0xe000
    movw r9, #4
    movt r9, #0x05fa
    str.w r9, [r8]
    dsb

    // Remain in SRAM if the failure reset is delayed.
commit_failure_reset_terminal:
    b.w commit_failure_reset_terminal
.size commit_engine, .-commit_engine

.type commit_fail_reset,%function
.thumb_func
commit_fail_reset:
    // Wait for an interrupted flash operation while feeding the watchdog.
commit_fail_reset_busy:
    movw r10, #0x3c0c
    movt r10, #0x4002
    movw r11, #0x3000
    movt r11, #0x4000
    movw r12, #0xaaaa
    movt r12, #0
    movw r5, #0
    movt r5, #1
commit_fail_reset_wait:
    ldr.w r6, [r10]
    str.w r12, [r11]
    tst r6, r5
    bne.w commit_fail_reset_wait

    // Lock the controller and request reset from SRAM.
commit_fail_reset_request:
    movw r8, #0x3c10
    movt r8, #0x4002
    movw r9, #0
    movt r9, #0x8000
    str.w r9, [r8]
    movw r8, #0xed0c
    movt r8, #0xe000
    movw r9, #4
    movt r9, #0x05fa
    str.w r9, [r8]

    // Remain in SRAM if the handler reset is delayed.
commit_fail_reset_terminal:
    b.w commit_fail_reset_terminal
.size commit_fail_reset, .-commit_fail_reset

.global commit_fault_poison
.hidden commit_fault_poison
.set commit_fault_poison, commit_poison_activation
