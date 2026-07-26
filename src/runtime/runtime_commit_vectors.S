.syntax unified
.cpu cortex-m4
.thumb
.eabi_attribute 7, 77

.section .ramvectors.commit,"aw",%progbits
.global commit_vectors
.type commit_vectors,%object
commit_vectors:
    .rept 128
    .word commit_fail_reset
    .endr
.size commit_vectors, .-commit_vectors

.global commit_poison_vectors
.type commit_poison_vectors,%object
commit_poison_vectors:
    .rept 128
    .word commit_fault_poison + 1
    .endr
.size commit_poison_vectors, .-commit_poison_vectors
