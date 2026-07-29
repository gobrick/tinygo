.syntax unified
.cpu cortex-m4
.thumb
.eabi_attribute 7, 77

.section .text.kinetigo_stop_ports,"ax",%progbits
.balign 4

.global kinetigo_stop_ports
.type kinetigo_stop_ports,%function
.thumb_func
kinetigo_stop_ports:
    movw r0, #0x3830
    movt r0, #0x4002
    ldr r1, [r0]
    orr r1, r1, #0x0f
    str r1, [r0]

stop_ports_coast:
    movw r0, #0x0400
    movt r0, #0x4002
    movw r1, #0
    movt r1, #0x0300
    str r1, [r0, #0x18]
    ldr r1, [r0]
    movw r2, #0
    movt r2, #0x000f
    bic r1, r1, r2
    movw r2, #0
    movt r2, #0x0005
    orr r1, r1, r2
    str r1, [r0]
    ldr r1, [r0, #0x04]
    bic r1, r1, #0x0300
    str r1, [r0, #0x04]
    ldr r1, [r0, #0x0c]
    movw r2, #0
    movt r2, #0x000f
    bic r1, r1, r2
    str r1, [r0, #0x0c]

stop_ports_timer:
    movw r0, #0x0800
    movt r0, #0x4000
    movs r1, #0
    str r1, [r0]
    str r1, [r0, #0x0c]
    str r1, [r0, #0x10]
    str r1, [r0, #0x20]
    movw r0, #0x3800
    movt r0, #0x4002
    ldr r1, [r0, #0x20]
    orr r1, r1, #0x04
    str r1, [r0, #0x20]
    bic r1, r1, #0x04
    str r1, [r0, #0x20]
    ldr r1, [r0, #0x40]
    bic r1, r1, #0x04
    str r1, [r0, #0x40]

stop_ports_buffer:
    movw r0, #0x0400
    movt r0, #0x4002
    movs r1, #0x04
    str r1, [r0, #0x18]
    ldr r1, [r0]
    bic r1, r1, #0x30
    orr r1, r1, #0x10
    str r1, [r0]
    ldr r1, [r0, #0x04]
    bic r1, r1, #0x04
    str r1, [r0, #0x04]
    ldr r1, [r0, #0x0c]
    bic r1, r1, #0x30
    str r1, [r0, #0x0c]

stop_ports_uart:
    movw r0, #0x5000
    movt r0, #0x4000
    movs r1, #0
    str r1, [r0, #0x0c]
    str r1, [r0, #0x10]
    str r1, [r0, #0x14]
    movw r0, #0x3800
    movt r0, #0x4002
    ldr r1, [r0, #0x20]
    movw r2, #0
    movt r2, #0x0010
    orr r1, r1, r2
    str r1, [r0, #0x20]
    bic r1, r1, r2
    str r1, [r0, #0x20]
    ldr r1, [r0, #0x40]
    bic r1, r1, r2
    str r1, [r0, #0x40]

stop_ports_uart_gpio:
    movw r0, #0x0800
    movt r0, #0x4002
    ldr r1, [r0]
    movw r2, #0
    movt r2, #0x0300
    bic r1, r1, r2
    str r1, [r0]
    ldr r1, [r0, #0x0c]
    bic r1, r1, r2
    str r1, [r0, #0x0c]
    ldr r1, [r0, #0x24]
    movw r2, #0
    movt r2, #0x000f
    bic r1, r1, r2
    str r1, [r0, #0x24]

    movw r0, #0x0c00
    movt r0, #0x4002
    ldr r1, [r0]
    bic r1, r1, #0x30
    str r1, [r0]
    ldr r1, [r0, #0x0c]
    bic r1, r1, #0x30
    str r1, [r0, #0x0c]
    ldr r1, [r0, #0x20]
    movw r2, #0x0f00
    bic r1, r1, r2
    str r1, [r0, #0x20]

stop_ports_vcc:
    movw r0, #0
    movt r0, #0x4002
    movw r1, #0
    movt r1, #0x4000
    str r1, [r0, #0x18]
    ldr r1, [r0]
    movw r2, #0
    movt r2, #0x3000
    bic r1, r1, r2
    movw r2, #0
    movt r2, #0x1000
    orr r1, r1, r2
    str r1, [r0]
    ldr r1, [r0, #0x0c]
    movw r2, #0
    movt r2, #0x3000
    bic r1, r1, r2
    str r1, [r0, #0x0c]
    bx lr

.size kinetigo_stop_ports, .-kinetigo_stop_ports

.section .text.kinetigo_hardfault_entry,"ax",%progbits
.balign 2

.global kinetigo_hardfault_entry
.type kinetigo_hardfault_entry,%function
.thumb_func
kinetigo_hardfault_entry:
    mov r12, lr
    bl kinetigo_stop_ports
    mov lr, r12
    b HardFault_Handler
.size kinetigo_hardfault_entry, .-kinetigo_hardfault_entry
