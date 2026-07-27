/* Stock frozen loader jumps to 0x08008000 (S2 vector-table probe, 2026-07-17;
 * matches spike-rt deploy address). Sector 11 stays reserved for breadcrumbs. */
flash_origin = DEFINED(flash_origin) ? flash_origin : 0x08008000;
flash_end = DEFINED(flash_end) ? flash_end : 0x080e0000;

MEMORY
{
    FLASH_TEXT (rw) : ORIGIN = flash_origin, LENGTH = flash_end - flash_origin
    RAM (xrw)       : ORIGIN = 0x20000000, LENGTH = 320K
}

_stack_size = 4K;

ENTRY(Reset_Handler)

SECTIONS
{
    .text :
    {
        KEEP(*(.isr_vector))
        KEEP(*(.after_isr_vector))
        ASSERT(. <= ORIGIN(FLASH_TEXT) + 0x1f0, "vector table overlaps firmware info")
        . = ORIGIN(FLASH_TEXT) + 0x1f0;
        KEEP(*(.fw_info))
        ASSERT(. <= ORIGIN(FLASH_TEXT) + 0x264, "first firmware info overlaps second")
        . = ORIGIN(FLASH_TEXT) + 0x264;
        KEEP(*(.fw_info2))
        *(.text)
        *(.text.*)
        *(.rodata)
        *(.rodata.*)
        . = ALIGN(4);
    } >FLASH_TEXT

    .tinygo_stacksizes : { *(.tinygo_stacksizes) } >FLASH_TEXT

    .stack (NOLOAD) :
    {
        . = ALIGN(4);
        . += _stack_size;
        _stack_top = .;
    } >RAM

    .stack1 (NOLOAD) :
    {
        . = ALIGN(4);
        . += DEFINED(__num_stacks) && __num_stacks >= 2 ? _stack_size : 0;
        _stack1_top = .;
    } >RAM

    .ramvectors.commit ALIGN(512) :
    {
        _sdata = .;
        __commit_vectors_start = .;
        KEEP(*(.ramvectors.commit))
        __commit_vectors_end = .;
    } >RAM AT>FLASH_TEXT

    .ramfuncs.commit ALIGN(4) :
    {
        __commit_code_start = .;
        KEEP(*(.ramfuncs.commit))
        __commit_code_end = .;
    } >RAM AT>FLASH_TEXT

    .data :
    {
        . = ALIGN(4);
        *(.data)
        *(.data.*)
        . = ALIGN(4);
        *(.ramfuncs*)
        . = ALIGN(4);
        _edata = .;
    } >RAM AT>FLASH_TEXT

    .bss :
    {
        . = ALIGN(4);
        _sbss = .;
        *(.bss)
        *(.bss.*)
        *(COMMON)
        . = ALIGN(4);
        _ebss = .;
    } >RAM

    /DISCARD/ :
    {
        *(.ARM.exidx)
        *(.ARM.exidx.*)
    }
}

_sidata = LOADADDR(.ramvectors.commit);
__commit_vectors_load_start = LOADADDR(.ramvectors.commit);
__commit_code_load_start = LOADADDR(.ramfuncs.commit);
_heap_start = _ebss;
_heap_end = ORIGIN(RAM) + LENGTH(RAM);
_globals_start = ADDR(.data);
_globals_end = _ebss;
__flash_data_start = LOADADDR(.data) + SIZEOF(.data);
__flash_data_end = ORIGIN(FLASH_TEXT) + LENGTH(FLASH_TEXT);
