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

.macro POISON_VECTORS name
.global commit_poison_vectors_\name
.type commit_poison_vectors_\name,%object
commit_poison_vectors_\name:
    .rept 128
    .word commit_fault_poison_\name
    .endr
.size commit_poison_vectors_\name, .-commit_poison_vectors_\name
.endm

POISON_VECTORS launcher
POISON_VECTORS single0
POISON_VECTORS single1
POISON_VECTORS single2
POISON_VECTORS single3
POISON_VECTORS single4
POISON_VECTORS single5
POISON_VECTORS double0
POISON_VECTORS double1
POISON_VECTORS double2
POISON_VECTORS double3
POISON_VECTORS double4
