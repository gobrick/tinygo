.syntax unified
.cpu cortex-m4
.thumb
.eabi_attribute 7, 77

.section .probe_vectors,"a",%progbits
.global probe_vectors
probe_vectors:
    .word probe_stack
    .word probe_reset

.section .probe_code,"ax",%progbits
.thumb_func
.global probe_reset
probe_reset:
    cpsid i

    ldr r0, =0x40023800
    ldr r1, [r0, #0x30]
    orr r1, r1, #3
    str r1, [r0, #0x30]
    ldr r1, [r0, #0x40]
    orr r1, r1, #0x40
    str r1, [r0, #0x40]
    ldr r1, [r0, #0x44]
    orr r1, r1, #0x1000
    str r1, [r0, #0x44]

    ldr r1, [r0, #0x20]
    orr r1, r1, #0x40
    str r1, [r0, #0x20]
    bic r1, r1, #0x40
    str r1, [r0, #0x20]
    ldr r1, [r0, #0x24]
    orr r1, r1, #0x1000
    str r1, [r0, #0x24]
    bic r1, r1, #0x1000
    str r1, [r0, #0x24]

    ldr r0, =0x40020000
    ldr r1, [r0]
    ldr r2, =0xc000f800
    bic r1, r1, r2
    ldr r2, =0x40008800
    orr r1, r1, r2
    str r1, [r0]
    ldr r1, [r0, #4]
    ldr r2, =0x80a0
    bic r1, r1, r2
    str r1, [r0, #4]
    ldr r1, [r0, #8]
    ldr r2, =0xc000cc00
    orr r1, r1, r2
    str r1, [r0, #8]
    ldr r1, [r0, #0xc]
    ldr r2, =0xc000cc00
    bic r1, r1, r2
    str r1, [r0, #0xc]
    ldr r1, =0x50500000
    str r1, [r0, #0x20]
    ldr r1, =0x80000000
    str r1, [r0, #0x18]

    ldr r0, =0x40020400
    ldr r1, [r0]
    ldr r2, =0xc0000000
    bic r1, r1, r2
    ldr r2, =0x80000000
    orr r1, r1, r2
    str r1, [r0]
    ldr r1, [r0, #4]
    bic r1, r1, #0x8000
    str r1, [r0, #4]
    ldr r1, [r0, #8]
    orr r1, r1, #0xc0000000
    str r1, [r0, #8]
    ldr r1, [r0, #0xc]
    bic r1, r1, #0xc0000000
    str r1, [r0, #0xc]
    ldr r1, =0x90000000
    str r1, [r0, #0x24]

    ldr r0, =0x40001800
    movs r1, #0
    str r1, [r0, #0x00]
    str r1, [r0, #0x0c]
    str r1, [r0, #0x10]
    str r1, [r0, #0x20]
    str r1, [r0, #0x28]
    movs r1, #1
    str r1, [r0, #0x2c]
    str r1, [r0, #0x38]
    ldr r1, =0x6000
    str r1, [r0, #0x18]
    movs r1, #0x10
    str r1, [r0, #0x20]
    movs r1, #1
    str r1, [r0, #0x14]
    movs r1, #0x81
    str r1, [r0]

    ldr r0, =0x40013000
    movs r1, #0
    str r1, [r0, #4]
    ldr r1, =0x344
    str r1, [r0]

    ldr r2, =probe_control
    movs r3, #97
    bl probe_send
    ldr r2, =probe_control
    movs r3, #97
    bl probe_send
    ldr r2, =probe_red
    movs r3, #97
    bl probe_send

probe_hold:
    b probe_hold

.thumb_func
probe_send:
    ldr r0, =0x40013000
probe_txe:
    ldr r1, [r0, #8]
    tst r1, #2
    beq probe_txe
    ldrb r1, [r2], #1
    strb r1, [r0, #0xc]
    subs r3, r3, #1
    bne probe_txe
probe_done:
    ldr r1, [r0, #8]
    tst r1, #2
    beq probe_done
probe_busy:
    ldr r1, [r0, #8]
    tst r1, #0x80
    bne probe_busy
    ldr r1, [r0, #0xc]
    ldr r1, [r0, #8]
    ldr r0, =0x40020000
    movw r1, #0x8000
    str r1, [r0, #0x18]
    nop
    nop
    nop
    nop
    nop
    nop
    nop
    nop
    nop
    nop
    nop
    nop
    nop
    nop
    nop
    nop
    movs r1, #0
    movt r1, #0x8000
    str r1, [r0, #0x18]
    bx lr

.balign 4
probe_control:
    .byte 0x01, 0x96
    .zero 48
    .byte 0x06, 0x7f, 0xff, 0xfe, 0
    .rept 42
    .byte 0xff
    .endr
probe_red:
    .zero 7
    .byte 0xff, 0xff
    .zero 4
    .byte 0xff, 0xff
    .zero 14
    .byte 0xff, 0xff
    .zero 66

.ltorg
