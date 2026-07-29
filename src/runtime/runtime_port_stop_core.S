.syntax unified
.cpu cortex-m4
.thumb
.eabi_attribute 7, 77

.section .text.kinetigo_stop_ports,"ax",%progbits
.balign 2

.global kinetigo_stop_ports
.type kinetigo_stop_ports,%function
.thumb_func
kinetigo_stop_ports:
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
