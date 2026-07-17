/* mboot probes two LEGO identity records (pybricks@76bfc071: platform.c). */
.section .fw_info,"a",%progbits
.global fw_info
.type fw_info,%object
fw_info:
.byte 0x0d
.ascii "STM32F413"
.zero 6
.byte 0
.ascii "SPIKE Prime"
.zero 4
.size fw_info, .-fw_info

.section .fw_info2,"a",%progbits
.global fw_info2
.type fw_info2,%object
fw_info2:
.byte 0x0d
.ascii "STM32F413"
.zero 6
.byte 0
.ascii "SPIKE Prime"
.zero 4
.size fw_info2, .-fw_info2
