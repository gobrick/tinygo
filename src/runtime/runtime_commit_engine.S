.syntax unified
.cpu cortex-m4
.thumb
.eabi_attribute 7, 77

.section .ramfuncs.commit,"ax",%progbits
.balign 4

.macro LOAD32 register, low, high
    movw \register, #\low
    movt \register, #\high
.endm

.macro SELECT_PLAN name, base_high, capacity_low, capacity_high, vector_literal, erase_label
commit_plan_\name:
    // Load this plan's constant base, capacity, and poison table.
    LOAD32 r3, 0, \base_high
    LOAD32 r7, \capacity_low, \capacity_high
    cmp r1, r7
    bhi.w commit_failure_reset
    ldr.w r14, \vector_literal
    b.w \erase_label
.endm

.macro ERASE plan, sector, next
commit_\plan\()_erase_\sector:
    // Admit and start this plan's fixed sector erase.
    LOAD32 r8, 0x7004, 0x4000
    ldr.w r6, [r8]
    LOAD32 r5, 4, 0
    tst r6, r5
    bne.w commit_failure_reset
    LOAD32 r8, 0x3c0c, 0x4002
    LOAD32 r9, 0x01f3, 0
    str.w r9, [r8]
    LOAD32 r8, 0x3c10, 0x4002
    LOAD32 r9, (2 | (\sector << 3)), 1
    str.w r9, [r8]

commit_\plan\()_erase_\sector\()_busy:
    // Wait for this plan's fixed sector erase.
    LOAD32 r10, 0x3c0c, 0x4002
    LOAD32 r11, 0x3000, 0x4000
    LOAD32 r12, 0xaaaa, 0
    LOAD32 r5, 0, 1
commit_\plan\()_erase_\sector\()_wait:
    ldr.w r6, [r10]
    str.w r12, [r11]
    tst r6, r5
    bne.w commit_\plan\()_erase_\sector\()_wait

commit_\plan\()_erase_\sector\()_status:
    // Clear this erase and reject every error flag.
    LOAD32 r8, 0x3c10, 0x4002
    LOAD32 r9, 0, 0
    str.w r9, [r8]
    LOAD32 r8, 0x3c0c, 0x4002
    ldr.w r6, [r8]
    LOAD32 r5, 0x01f2, 0
    tst r6, r5
    bne.w commit_failure_reset
    b.w \next
.endm

.global commit_engine
.global commit_fail_reset

.type commit_engine,%function
.thumb_func
commit_engine:
    // Mask configurable interrupts and stop SysTick.
commit_mask_interrupts:
    cpsid i
    LOAD32 r8, 0xe010, 0xe000
    LOAD32 r9, 0, 0
    str.w r9, [r8]

commit_disable_interrupts:
    // Disable every configurable interrupt.
    LOAD32 r8, 0xe180, 0xe000
    LOAD32 r9, 0xffff, 0xffff
    str.w r9, [r8, #0x00]
    str.w r9, [r8, #0x04]
    str.w r9, [r8, #0x08]
    str.w r9, [r8, #0x0c]
    str.w r9, [r8, #0x10]
    str.w r9, [r8, #0x14]
    str.w r9, [r8, #0x18]
    str.w r9, [r8, #0x1c]

commit_clear_interrupts:
    // Clear every pending configurable interrupt.
    LOAD32 r8, 0xe280, 0xe000
    str.w r9, [r8, #0x00]
    str.w r9, [r8, #0x04]
    str.w r9, [r8, #0x08]
    str.w r9, [r8, #0x0c]
    str.w r9, [r8, #0x10]
    str.w r9, [r8, #0x14]
    str.w r9, [r8, #0x18]
    str.w r9, [r8, #0x1c]

commit_install_vectors:
    // Install and synchronize the initial SRAM vector table.
    LOAD32 r8, 0xed08, 0xe000
    LOAD32 r9, 0x1000, 0x2000
    str.w r9, [r8]
    dsb
    isb

commit_reset_caches:
    // Disable and reset both flash caches.
    LOAD32 r8, 0x3c00, 0x4002
    LOAD32 r9, 0, 0
    str.w r9, [r8]
    LOAD32 r9, 0x1800, 0
    str.w r9, [r8]
    LOAD32 r9, 0, 0
    str.w r9, [r8]

commit_initial_busy:
    // Wait for a prior flash operation while feeding the watchdog.
    LOAD32 r10, 0x3c0c, 0x4002
    LOAD32 r11, 0x3000, 0x4000
    LOAD32 r12, 0xaaaa, 0
    LOAD32 r5, 0, 1
commit_initial_wait:
    ldr.w r6, [r10]
    str.w r12, [r11]
    tst r6, r5
    bne.w commit_initial_wait

commit_unlock_flash:
    // Unlock the flash controller with the fixed key sequence.
    LOAD32 r8, 0x3c04, 0x4002
    LOAD32 r9, 0x0123, 0x4567
    str.w r9, [r8]
    LOAD32 r9, 0x89ab, 0xcdef
    str.w r9, [r8]
    LOAD32 r8, 0x3c0c, 0x4002
    ldr.w r6, [r8]

commit_bound_length:
    // Prove the admitted source length is inside the global receive bounds.
    LOAD32 r7, 4, 0
    cmp r1, r7
    blo.w commit_failure_reset
    LOAD32 r7, 0, 4
    cmp r1, r7
    bhi.w commit_failure_reset

commit_bound_source:
    // Prove the admitted source span is inside SRAM.
    LOAD32 r7, 0, 0x2000
    cmp r0, r7
    blo.w commit_failure_reset
    LOAD32 r7, 0, 0x2005
    subs r7, r7, r1
    cmp r0, r7
    bhi.w commit_failure_reset

commit_select_plan:
    // Range-check the selector through immediate comparisons only.
    cmp r2, #0
    beq.w commit_plan_launcher
    cmp r2, #1
    beq.w commit_plan_single0
    cmp r2, #2
    beq.w commit_plan_single1
    cmp r2, #3
    beq.w commit_plan_single2
    cmp r2, #4
    beq.w commit_plan_single3
    cmp r2, #5
    beq.w commit_plan_single4
    cmp r2, #6
    beq.w commit_plan_single5
    cmp r2, #7
    beq.w commit_plan_double0
    cmp r2, #8
    beq.w commit_plan_double1
    cmp r2, #9
    beq.w commit_plan_double2
    cmp r2, #10
    beq.w commit_plan_double3
    cmp r2, #11
    beq.w commit_plan_double4
    b.w commit_failure_reset

SELECT_PLAN launcher, 0x0801, 0x0000, 1, commit_plan_launcher_vector_literal, commit_launcher_erase_4
SELECT_PLAN single0, 0x0802, 0x0000, 2, commit_plan_single0_vector_literal, commit_single0_erase_5
SELECT_PLAN single1, 0x0804, 0x0000, 2, commit_plan_single1_vector_literal, commit_single1_erase_6
SELECT_PLAN single2, 0x0806, 0x0000, 2, commit_plan_single2_vector_literal, commit_single2_erase_7
SELECT_PLAN single3, 0x0808, 0x0000, 2, commit_plan_single3_vector_literal, commit_single3_erase_8
SELECT_PLAN single4, 0x080a, 0x0000, 2, commit_plan_single4_vector_literal, commit_single4_erase_9
SELECT_PLAN single5, 0x080c, 0x0000, 2, commit_plan_single5_vector_literal, commit_single5_erase_10
SELECT_PLAN double0, 0x0802, 0x0000, 4, commit_plan_double0_vector_literal, commit_double0_erase_5
SELECT_PLAN double1, 0x0804, 0x0000, 4, commit_plan_double1_vector_literal, commit_double1_erase_6
SELECT_PLAN double2, 0x0806, 0x0000, 4, commit_plan_double2_vector_literal, commit_double2_erase_7
SELECT_PLAN double3, 0x0808, 0x0000, 4, commit_plan_double3_vector_literal, commit_double3_erase_8
SELECT_PLAN double4, 0x080a, 0x0000, 4, commit_plan_double4_vector_literal, commit_double4_erase_9

ERASE launcher, 4, commit_blank_initialize
ERASE single0, 5, commit_blank_initialize
ERASE single1, 6, commit_blank_initialize
ERASE single2, 7, commit_blank_initialize
ERASE single3, 8, commit_blank_initialize
ERASE single4, 9, commit_blank_initialize
ERASE single5, 10, commit_blank_initialize
ERASE double0, 5, commit_double0_erase_6
ERASE double0, 6, commit_blank_initialize
ERASE double1, 6, commit_double1_erase_7
ERASE double1, 7, commit_blank_initialize
ERASE double2, 7, commit_double2_erase_8
ERASE double2, 8, commit_blank_initialize
ERASE double3, 8, commit_double3_erase_9
ERASE double3, 9, commit_blank_initialize
ERASE double4, 9, commit_double4_erase_10
ERASE double4, 10, commit_blank_initialize

commit_blank_initialize:
    // Initialize the full-plan blank verification.
    LOAD32 r2, 0, 0

commit_blank_loop:
    // Verify every erased byte while feeding the watchdog.
    cmp r2, r7
    bhs.w commit_after_blank
    LOAD32 r8, 0x3000, 0x4000
    LOAD32 r9, 0xaaaa, 0
    str.w r9, [r8]
    ldrb r4, [r3, r2]
    cmp r4, #0xff
    bne.w commit_failure_reset
    adds r2, #1
    b.w commit_blank_loop

commit_after_blank:
    // Initialize byte programming after the blank span passes.
    LOAD32 r2, 0, 0

commit_program_loop:
    // Skip activation and erased source bytes in the body loop.
    cmp r2, r1
    bhs.w commit_after_program
    LOAD32 r8, 0x3000, 0x4000
    LOAD32 r9, 0xaaaa, 0
    str.w r9, [r8]
    cmp r2, #3
    beq.w commit_program_increment
    ldrb r4, [r0, r2]
    cmp r4, #0xff
    beq.w commit_program_increment

commit_program_byte:
    // Admit and arm one x8 body-byte program.
    LOAD32 r8, 0x7004, 0x4000
    ldr.w r6, [r8]
    LOAD32 r5, 4, 0
    tst r6, r5
    bne.w commit_failure_reset
    LOAD32 r8, 0x3c0c, 0x4002
    LOAD32 r9, 0x01f3, 0
    str.w r9, [r8]
    LOAD32 r8, 0x3c10, 0x4002
    LOAD32 r9, 1, 0
    str.w r9, [r8]
    strb r4, [r3, r2]

commit_program_byte_busy:
    // Wait for one body byte while feeding the watchdog.
    LOAD32 r10, 0x3c0c, 0x4002
    LOAD32 r11, 0x3000, 0x4000
    LOAD32 r12, 0xaaaa, 0
    LOAD32 r5, 0, 1
commit_program_byte_wait:
    ldr.w r6, [r10]
    str.w r12, [r11]
    tst r6, r5
    bne.w commit_program_byte_wait

commit_program_byte_status:
    // Clear body-program control bits and reject every error flag.
    LOAD32 r8, 0x3c10, 0x4002
    LOAD32 r9, 0, 0
    str.w r9, [r8]
    LOAD32 r8, 0x3c0c, 0x4002
    ldr.w r6, [r8]
    LOAD32 r5, 0x01f2, 0
    tst r6, r5
    bne.w commit_failure_reset

commit_program_increment:
    // Advance the body-program induction variable once.
    adds r2, #1
    b.w commit_program_loop

commit_after_program:
    // Initialize the pre-activation image and tail comparison.
    LOAD32 r2, 0, 0

commit_preverify_loop:
    // Compare the body and erased tail before activation.
    cmp r2, r7
    bhs.w commit_activate
    LOAD32 r8, 0x3000, 0x4000
    LOAD32 r9, 0xaaaa, 0
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

commit_preverify_tail:
    // Require erased flash at activation and beyond the image.
    cmp r5, #0xff
    bne.w commit_failure_reset

commit_preverify_increment:
    // Advance the pre-activation comparison once.
    adds r2, #1
    b.w commit_preverify_loop

commit_activate:
    // Install this plan's fault-to-poison vector table.
    LOAD32 r8, 0xed08, 0xe000
    str.w r14, [r8]
    dsb
    isb

commit_activate_write:
    // Admit and program this plan's activation byte last.
    LOAD32 r8, 0x7004, 0x4000
    ldr.w r6, [r8]
    LOAD32 r5, 4, 0
    tst r6, r5
    bne.w commit_failure_reset
    LOAD32 r8, 0x3c0c, 0x4002
    LOAD32 r9, 0x01f3, 0
    str.w r9, [r8]
    LOAD32 r8, 0x3c10, 0x4002
    LOAD32 r9, 1, 0
    str.w r9, [r8]
    LOAD32 r4, 0x20, 0
    strb r4, [r3, #3]

commit_activate_busy:
    // Wait for activation while feeding the watchdog.
    LOAD32 r10, 0x3c0c, 0x4002
    LOAD32 r11, 0x3000, 0x4000
    LOAD32 r12, 0xaaaa, 0
    LOAD32 r5, 0, 1
commit_activate_wait:
    ldr.w r6, [r10]
    str.w r12, [r11]
    tst r6, r5
    bne.w commit_activate_wait

commit_activate_status:
    // Clear activation control bits and poison on any error flag.
    LOAD32 r8, 0x3c10, 0x4002
    LOAD32 r9, 0, 0
    str.w r9, [r8]
    LOAD32 r8, 0x3c0c, 0x4002
    ldr.w r6, [r8]
    LOAD32 r5, 0x01f2, 0
    tst r6, r5
    bne.w commit_poison_activation

commit_postverify_initialize:
    // Initialize the post-activation image and tail comparison.
    LOAD32 r2, 0, 0

commit_postverify_loop:
    // Compare the complete image and erased tail after activation.
    cmp r2, r7
    bhs.w commit_success_reset
    LOAD32 r8, 0x3000, 0x4000
    LOAD32 r9, 0xaaaa, 0
    str.w r9, [r8]
    ldrb r5, [r3, r2]
    cmp r2, r1
    bhs.w commit_postverify_tail
    ldrb r4, [r0, r2]
    cmp r5, r4
    bne.w commit_poison_activation
    b.w commit_postverify_increment

commit_postverify_tail:
    // Require erased flash beyond the activated image.
    cmp r5, #0xff
    bne.w commit_poison_activation

commit_postverify_increment:
    // Advance the post-activation comparison once.
    adds r2, #1
    b.w commit_postverify_loop

commit_success_reset:
    // Lock the controller and request reset after full verification.
    LOAD32 r8, 0x3c10, 0x4002
    LOAD32 r9, 0, 0x8000
    str.w r9, [r8]
    LOAD32 r8, 0xed0c, 0xe000
    LOAD32 r9, 4, 0x05fa
    str.w r9, [r8]
    dsb

commit_success_reset_terminal:
    // Remain in SRAM if the successful reset is delayed.
    b.w commit_success_reset_terminal

commit_poison_activation:
    // Wait for any interrupted operation before poisoning.
    LOAD32 r10, 0x3c0c, 0x4002
    LOAD32 r11, 0x3000, 0x4000
    LOAD32 r12, 0xaaaa, 0
    LOAD32 r5, 0, 1
commit_poison_entry_wait:
    ldr.w r6, [r10]
    str.w r12, [r11]
    tst r6, r5
    bne.w commit_poison_entry_wait

commit_poison_supply:
    // Feed and retry until poisoning can start on acceptable supply.
    LOAD32 r8, 0x3000, 0x4000
    LOAD32 r9, 0xaaaa, 0
    str.w r9, [r8]
    LOAD32 r8, 0x7004, 0x4000
    ldr.w r6, [r8]
    LOAD32 r5, 4, 0
    tst r6, r5
    bne.w commit_poison_activation

commit_poison_write:
    // Program zero over this plan's poison address.
    LOAD32 r8, 0x3c0c, 0x4002
    LOAD32 r9, 0x01f3, 0
    str.w r9, [r8]
    LOAD32 r8, 0x3c10, 0x4002
    LOAD32 r9, 1, 0
    str.w r9, [r8]
    LOAD32 r4, 0, 0
    strb r4, [r3, #3]

commit_poison_busy:
    // Wait for poisoning while feeding the watchdog.
    LOAD32 r10, 0x3c0c, 0x4002
    LOAD32 r11, 0x3000, 0x4000
    LOAD32 r12, 0xaaaa, 0
    LOAD32 r5, 0, 1
commit_poison_wait:
    ldr.w r6, [r10]
    str.w r12, [r11]
    tst r6, r5
    bne.w commit_poison_wait

commit_poison_status:
    // Clear poison control bits and retry on any error flag.
    LOAD32 r8, 0x3c10, 0x4002
    LOAD32 r9, 0, 0
    str.w r9, [r8]
    LOAD32 r8, 0x3c0c, 0x4002
    ldr.w r6, [r8]
    LOAD32 r5, 0x01f2, 0
    tst r6, r5
    bne.w commit_poison_activation

commit_poison_verify:
    // Reread until activation is no longer live.
    ldrb r4, [r3, #3]
    cmp r4, #0x20
    beq.w commit_poison_activation
    b.w commit_failure_reset

commit_failure_reset:
    // Lock the controller and request reset on a safe failure path.
    LOAD32 r8, 0x3c10, 0x4002
    LOAD32 r9, 0, 0x8000
    str.w r9, [r8]
    LOAD32 r8, 0xed0c, 0xe000
    LOAD32 r9, 4, 0x05fa
    str.w r9, [r8]
    dsb

commit_failure_reset_terminal:
    // Remain in SRAM if the failure reset is delayed.
    b.n commit_failure_reset_terminal

.balign 4
commit_plan_launcher_vector_literal:
    .word commit_poison_vectors_launcher
commit_plan_single0_vector_literal:
    .word commit_poison_vectors_single0
commit_plan_single1_vector_literal:
    .word commit_poison_vectors_single1
commit_plan_single2_vector_literal:
    .word commit_poison_vectors_single2
commit_plan_single3_vector_literal:
    .word commit_poison_vectors_single3
commit_plan_single4_vector_literal:
    .word commit_poison_vectors_single4
commit_plan_single5_vector_literal:
    .word commit_poison_vectors_single5
commit_plan_double0_vector_literal:
    .word commit_poison_vectors_double0
commit_plan_double1_vector_literal:
    .word commit_poison_vectors_double1
commit_plan_double2_vector_literal:
    .word commit_poison_vectors_double2
commit_plan_double3_vector_literal:
    .word commit_poison_vectors_double3
commit_plan_double4_vector_literal:
    .word commit_poison_vectors_double4
.size commit_engine, .-commit_engine

.macro POISON_HANDLER name, base_high
.global commit_fault_poison_\name
.type commit_fault_poison_\name,%function
.thumb_func
commit_fault_poison_\name:
    // Reload this plan's constant base before common poison code.
    LOAD32 r3, 0, \base_high
    b.w commit_poison_activation
.size commit_fault_poison_\name, .-commit_fault_poison_\name
.endm

POISON_HANDLER launcher, 0x0801
POISON_HANDLER single0, 0x0802
POISON_HANDLER single1, 0x0804
POISON_HANDLER single2, 0x0806
POISON_HANDLER single3, 0x0808
POISON_HANDLER single4, 0x080a
POISON_HANDLER single5, 0x080c
POISON_HANDLER double0, 0x0802
POISON_HANDLER double1, 0x0804
POISON_HANDLER double2, 0x0806
POISON_HANDLER double3, 0x0808
POISON_HANDLER double4, 0x080a

.type commit_fail_reset,%function
.thumb_func
commit_fail_reset:
    // Wait for an interrupted flash operation while feeding the watchdog.
    LOAD32 r10, 0x3c0c, 0x4002
    LOAD32 r11, 0x3000, 0x4000
    LOAD32 r12, 0xaaaa, 0
    LOAD32 r5, 0, 1
commit_fail_reset_wait:
    ldr.w r6, [r10]
    str.w r12, [r11]
    tst r6, r5
    bne.w commit_fail_reset_wait

commit_fail_reset_request:
    // Lock the controller and request reset from SRAM.
    LOAD32 r8, 0x3c10, 0x4002
    LOAD32 r9, 0, 0x8000
    str.w r9, [r8]
    LOAD32 r8, 0xed0c, 0xe000
    LOAD32 r9, 4, 0x05fa
    str.w r9, [r8]

commit_fail_reset_terminal:
    // Remain in SRAM if the handler reset is delayed.
    b.w commit_fail_reset_terminal
.size commit_fail_reset, .-commit_fail_reset

.global commit_fault_poison
.hidden commit_fault_poison
.set commit_fault_poison, commit_fault_poison_launcher
