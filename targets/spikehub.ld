/* Layout: pybricks/pybricks-micropython@76bfc071652a91c6b5d443dbfa55f1f0fa7933cd
 * lib/pbio/platform/prime_hub_f4/platform.ld. */
MEMORY
{
    FLASH_TEXT (rw) : ORIGIN = 0x08010000, LENGTH = 960K
    RAM (xrw)       : ORIGIN = 0x20000000, LENGTH = 320K
}

_stack_size = 4K;

INCLUDE "targets/arm.ld"
