MEMORY
{
    FLASH_TEXT (rw) : ORIGIN = 0x08008000, LENGTH = 864K
    RAM (xrw)       : ORIGIN = 0x20000000, LENGTH = 320K
}

_stack_size = 4K;

ENTRY(probe_reset)

SECTIONS
{
    .text :
    {
        KEEP(*(.probe_vectors))
        ASSERT(. <= ORIGIN(FLASH_TEXT) + 0x1f0, "probe overlaps firmware info")
        . = ORIGIN(FLASH_TEXT) + 0x1f0;
        KEEP(*(.fw_info))
        ASSERT(. <= ORIGIN(FLASH_TEXT) + 0x264, "firmware info overlaps")
        . = ORIGIN(FLASH_TEXT) + 0x264;
        KEEP(*(.fw_info2))
        KEEP(*(.probe_code))
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

    .data :
    {
        . = ALIGN(4);
        _sdata = .;
        *(.data)
        *(.data.*)
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
        *(.isr_vector)
        *(.after_isr_vector)
        *(.ARM.exidx)
        *(.ARM.exidx.*)
    }
}

_sidata = LOADADDR(.data);
_heap_start = _ebss;
_heap_end = ORIGIN(RAM) + LENGTH(RAM);
_globals_start = _sdata;
_globals_end = _ebss;
__flash_data_start = LOADADDR(.data) + SIZEOF(.data);
__flash_data_end = ORIGIN(FLASH_TEXT) + LENGTH(FLASH_TEXT);
