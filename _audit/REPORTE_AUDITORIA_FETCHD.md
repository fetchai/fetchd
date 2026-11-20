# INFORME DE AUDITORÍA — fetchd / red Fetch.ai (fetchhub-4)

**Fecha:** 2026-11-20 · **Commit auditado:** `fetchai/fetchd` `ae12a4fc8886b6bd45f4c70fb4973cd065462ad0` (= master público a la fecha, = release de producción `v0.15.1`)
**Método:** sin RPC — todo derivado de clones git, artefactos binarios de los propios repos, docs de upgrade on-chain y genesis exportado.

---

## VEREDICTO

**SÍ HAY VULNERABILIDAD CRÍTICA REAL EN EL PIN DE LA DEPENDENCIA WASM.**
El binario de producción de la red principal (fetchd v0.15.1, en ejecución en fetchhub-4 desde la altura 29,254,333) linka **`libwasmvm` compilado con Wasmer `7.4.2` + backend `singlepass`**, que está **dentro del rango vulnerable confirmado del advisory crítico de Wasmer GHSA-v2h8-j899-r676** (miscompilación de `if` tipado con block params y `else` omitido → corrupción del stack nativo → **escape del sandbox y ejecución de código nativo en el proceso del nodo**), y de su advisory de ecosistema **CWA-2026-006** («permite minting no autorizado de tokens nativos y pérdida permanente de fondos de usuarios»). La entrega de contratos es **permissionless** según el genesis de fetchhub-4 (`code_upload_access: Everybody`, `instantiate_default_permission: Everybody`). Cumple los 4 criterios (ver §2).

---

## 1. REQUISITO PREVIO — mapeo de la fundación y del despliegue (sin RPC)

### 1.1 Orgs/espacios de la fundación

| Espacio | Estado |
|---|---|
| GitHub org `fetchai` (~25 repos) | accesible vía `api.github.com`; listado completo obtenido |
| CosmWasm (upstream + forks privados de seguridad) | `CosmWasm/wasmvm`, `CosmWasm/wasmd`, `CosmWasm/cosmwasm`, `CosmWasm/advisories`, `wasmerio/wasmer` clonados |
| Documentación | `docs/` dentro del propio repo fetchd + `fetchai/genesis-fetchhub` (procedimientos de upgrade reales por red) |

### 1.2 Repos clonados (fuente de verdad) y commit exacto

| Repo | Commit | Rol |
|---|---|---|
| `fetchai/fetchd` | `ae12a4fc8886b6bd45f4c70fb4973cd065462ad0` (master, 2026-09-17; verificado vía API que es el master público actual — **no hay commit posterior ni release posterior**) | nodo en producción |
| `fetchai/priv_wasmvm_sec` | pin `d89faf180d3fb5de09b4ba59099844cc0a7a2632` (= tag `v3.0.8-rc.3`, verificado con `git rev-parse`) | fork privado de seguridad de wasmvm, **es la dependencia wasm real** |
| `fetchai/priv_wasmd_sec` | pin `ab17af8e72692541e5eb1b57c2d5480ed8084300` (= tag `v0.61.15-rc.3`, «bump rc.3») | fork privado de wasmd |
| `CosmWasm/wasmvm` | tags `v3.0.8` (2026-09-28), `v3.0.9` (2026-10-08) | upstream |
| `CosmWasm/wasmd` | tags `v0.61.15` (2026-09-28), `v0.61.16` (2026-10-09) | upstream |
| `CosmWasm/cosmwasm` | tags `v3.0.10-rc.3`, `v3.0.10` (2026-09-28), `v3.0.11` (2026-10-08) | VM Rust |
| `wasmerio/wasmer` | tags `v7.4.2` (2026-09-16), `v7.5.0` (2026-10-01) + security-advisories API | compilador wasm |
| `fetchai/genesis-fetchhub` | `cdc8ced` | genesis + guías de upgrade reales |
| `fetchai/cosmos-sdk` | tag `v0.20.0` (`5a24a33`) | fork del SDK (replace en go.mod:271) |
| `fetchai/cw-fetch-std` | `99e6ca2` | std de contratos del ecosistema |
| `fetchai/FIPs` | `43d6036` | FIPs |

### 1.3 Repos bloqueados / no accesibles

- `CosmWasm/priv_cosmwasm` → `fatal: repository not found` (404, privado real). Es la dependencia de `priv_wasmvm_sec` (`libwasmvm/Cargo.toml:69-74`). **No se declara cobertura completa sobre él**; su contenido queda caracterizado indirectamente por el tag `v3.0.10-rc.3` espejado en el repo público `CosmWasm/cosmwasm` (verificado: `git diff v3.0.10-rc.3 v3.0.10 -- packages/vm packages/std packages/crypto` = vacío, es decir el código VM del rc.3 == final 3.0.10).
- Los binarios de release no reproducibles localmente: `musl.Dockerfile` requiere Alpine/musl (no instalado en sandbox); se verificó en cambio el artefacto que los repos commitean.

### 1.4 Mapeo del despliegue leído del propio repo + web

- **Red live:** `fetchhub-4` (docs fetchd: Dorado/archivadas; `genesis-fetchhub/fetchhub-4/`).
- **Último upgrade en curso:** propuesta de governance #41 (`genesis-fetchhub/fetchhub-4/10-software-upgrade-v0.15.1.md`), altura **29,254,333**, «update of the wasmvm & wasmd libraries to the latest rc3 versions», **binary-only release** («no es posible construir el binario desde el código fuente público» — los binarios vienen del pipeline privado), artefactos `fetchd_v0.15.1-ae12a4f_static_linux_{x86_64,aarch64}.tgz` y docker `fetchai/fetchd:v0.15.1-ae12a4f-musl`. `app/upgrades.go:48` define `UpgradeNameV0_15_1 = "v0.15.1"` como el upgrade actual; no hay plan v0.16 en el código.
- **Cadena de dependencia wasm (verificada hash a hash):**
  1. `fetchd/go.mod:247-278` — `replace github.com/CosmWasm/wasmd => github.com/fetchai/priv_wasmd_sec v0.61.15-rc.3` y `replace github.com/CosmWasm/wasmvm/v3 => github.com/fetchai/priv_wasmvm_sec/v3 v3.0.8-rc.3` (comentario en go.mod con los commits exactos, que coinciden con los tags verificados).
  2. `priv_wasmvm_sec@d89faf1/libwasmvm/Cargo.toml:29-37` — `cosmwasm-vm = "=3.0.10-rc.3"`, `features = ["singlepass"]`; `Cargo.lock` — **`wasmer 7.4.2`** (todas las cajas wasmer-*).
  3. Los binarios **precompilados commiteados** en el módulo Go: `internal/api/libwasmvm_muslc.x86_64.a.xz` (blob `6afef242…`) y `internal/api/libwasmvm_muslc.aarch64.a.xz` (blob `9925d7c6…`).
  4. `fetchd/Makefile:171-173` — target de release `install-static-v0.15.1` fija `WASMVM_VERSION := v3.0.8-rc.3` y los sha256 de los artefactos **descomprimidos**: x86_64 `ab1a878ed3beecc5f3821b1fb1f73a5254443e3ba9bca53964a50a122ee77497`, aarch64 `cc9175235d1e0051002c33a94027a63a008b14dcc13ae483685e511d59754235`. **Verifiqué ambos sha256 localmente contra los blobs del fork: coinciden exactamente.** `docker-bake.hcl:48` y `musl.Dockerfile:108` usan ese target para las imágenes de release.
  5. Ejecutado `version_str()` sobre el `.so` extraído del commit pinneado → **`"3.0.8-rc.3"`**; `strings` del `.a` → símbolos `wasmer_compiler_singlepass` y literal `7.4.2`.

---

## 2. HALLAZGO CRÍTICO — Escape del sandbox de CosmWasm en el compilador Wasmer Singlepass (dependencia wasm desactualizada)

### 2.1 Ubicación exacta

| Elemento | Valor |
|---|---|
| Repo de la app | `fetchai/fetchd`, commit `ae12a4fc8886b6bd45f4c70fb4973cd065462ad0` (= release `v0.15.1` = master actual) |
| Declaración del pin | `go.mod:247-278` — `replace github.com/CosmWasm/wasmvm/v3 => github.com/fetchai/priv_wasmvm_sec/v3 v3.0.8-rc.3`; `go.sum:941` |
| Librería vulnerable | `priv_wasmvm_sec` @ `d89faf180d3fb5de09b4ba59099844cc0a7a2632`, `libwasmvm/Cargo.toml:35-37` (`cosmwasm-vm = "=3.0.10-rc.3"`, `wasmer = 7.4.2`, `features=["singlepass"]`), binario conmutado `internal/api/libwasmvm_muslc.x86_64.a.xz` / `libwasmvm.x86_64.so` |
| Selección del backend vulnerable | `CosmWasm/cosmwasm@v3.0.10` `packages/vm/src/wasm_backend/engine.rs:54` — `wasmer::sys::Singlepass::new()` (sin restricción de multi-value) |
| Defecto de codegen | `wasmerio/wasmer` `wasmer-compiler-singlepass`, introducido en PR #5869 (multi-value singlepass), presente en `>= 7.0.0-alpha.1` incluido **7.4.2** |
| Advisories | **GHSA-v2h8-j899-r676** (critical, 2026-11-02) y **CWA-2026-006** (critical, fund loss) |

### 2.2 Explicación del bug en código

Wasmer Singlepass lowered incorrectamente un `if` tipado **con block parameters y `else` omitido** (wasm válido): la etiqueta del brazo falso se vincula **después** del código de limpieza de los block parameters spill-eados al stack nativo. Con condición falsa, el salto se pasa esa limpieza y el modelo de stack del compilador asume que ocurrió: **cada ejecución afectada deja un slot de 8 bytes adicional en el stack nativo** (drift). Ejecuciones repetidas hacen drift del puntero de stack nativo respecto de la posición trackeada; los accesos nativos posteriores escriben **fuera del frame previsto**, lo que habilita **ejecución de código nativo con los privilegios del proceso del nodo** (y, como mínimo, crash del validador). El advisory de Wasmer lo declara explícitamente: *"The reporter demonstrated native code execution in a Linux x86-64 CosmWasm/wasmvm deployment. Applications exposing unpermissioned contract upload and execution can therefore be remotely exploitable."* El advisory de ecosistema CWA-2026-006 cuantifica el impacto: *"permits unauthorized minting of native tokens and permanent loss of user funds… The resulting balances are indistinguishable from ordinary ones and survive a node restart."*

Cronología verificada (relevante para entender por qué el pin «parcheado» sigue vulnerable):

1. **2026-08-18**: reporte original (CWA-2026-006) contra la línea wasmer 5.0.6 (wasmvm ≤ v3.0.7).
2. **2026-09-10/16**: parche privado = upgrade a Wasmer 7.4.2 (cosmwasm-vm 3.0.10-rc.3), distribuido por los forks privados; **`priv_wasmvm_sec v3.0.8-rc.3` (d89faf1, 2026-09-16) es exactamente ese parche**.
3. **2026-09-17**: fetchd **v0.15.1** adopta los rc.3 (commit `ae12a4f` = HEAD; «Update wasmvm & wasmd to the latest rc3») y se aplica on-chain (propuesta #41).
4. **2026-11-02**: **GHSA-v2h8** — la propia línea «parcheada» (wasmvm 3.0.8/2.3.5, wasmer 7.4.2) está afectada por un defecto de codegen singlepass presente en todo wasmer ≥ 7.0.0-alpha.1, ≤ 7.4.2. Fix preparado en rama privada; **versión parcheada sin confirmar; hasta hoy no existe release pública de wasmer con el fix** (`patched: NONE-YET`; el último advisory, 2026-11-13, sigue sin parche para < 7.6.0).
5. **2026-10-08/09**: CosmWasm publicó la línea de remediación siguiente (wasmvm **v3.0.9** = cosmwasm **3.0.11** = wasmer **7.5.0**; wasmd **v0.61.16**/v0.70.5). **fetchd no la adoptó**: su master (2026-09-17) y su último release siguen en rc.3/wasmer 7.4.2. Además, verifiqué el diff público wasmer `v7.4.2 → v7.5.0`: no contiene ningún cambio en el lowering de singlepass (solo validación de tablas fijas = GHSA-xmhf), i.e. ni siquiera la línea de remediación actual incorpora un fix público del defecto de lowering.

### 2.3 Condiciones de configuración (leídas del repo, sin RPC)

- **Entrega permissionless (criterio 2):** `fetchai/genesis-fetchhub`, `fetchhub-4/data/genesis_migrated_5300200.json` → `app_state.wasm.params`:
  ```json
  "code_upload_access": {"address": "", "permission": "Everybody"},
  "instantiate_default_permission": "Everybody",
  "max_wasm_code_size": "614400"
  ```
  Cualquier cuenta financiada puede ejecutar `MsgStoreCode` + `MsgInstantiateContract` + `MsgExecuteContract` sin governance, sin rol privilegiado, sin clave de validador. CWA-2026-006 lo confirma como condición de explotabilidad; incluso con parámetros restrictivos, *«A contract already stored on chain by an attacker remains a path»*.
- **Ruta de ejecución activa (criterio 3):** el binario de release v0.15.1 se construye con `make install-static-v0.15.1` (`docker-bake.hcl:48`, `musl.Dockerfile:108`), que extrae a `/lib` el artefacto conmutado `internal/api/libwasmvm_muslc.x86_64.a.xz` del módulo pinneado (sha256 verificado abajo) y lo linka con build tag `muslc`. Toda ejecución de contrato en fetchhub-4 pasa por este compilador.
- **Parámetros que habilitan la ruta del defecto:** cosmwasm-vm 3.0.10 habilita el backend singlepass sin la mitigación del advisory (deshabilitar multi-value en validación) — `engine.rs:54` crea `Singlepass::new()` sin restricciones; los block parameters con `if` de un brazo pasan la validación estática de producción (lo demostré: `store_code(checked=True)` aceptó mis módulos de prueba).

### 2.4 PoC local / trazado reproducible (criterio 4)

Artefacto exacto de producción: extraje `internal/api/libwasmvm.x86_64.so` del tarball de `priv_wasmvm_sec@d89faf1` (vía codeload), sha256 `0b18713c01c57a12d956614c98de6bd247d29b0649e7fa91a53b9444ce80ac32`; `version_str()` del binario → **`3.0.8-rc.3`**.

Harness (Python/ctypes, sin RPC): `_audit/poc/poc.py` — `init_cache` → `store_code(checked=True)` → `instantiate` → `execute`, con vtables Db/Api/Querier simuladas. Contratos sintetizados a mano (`build_wasm.py`, `build_straight.py`) que ejercitan el constructo del advisory: `if` tipado de un brazo (`(param i32) (result i32)`, params==results) con `else` omitido, condición falsa, valores externos vivos para forzar spill del block param.

Resultados ejecutados:

```
[CONTROL] else explícito, 2.000.000 iteraciones en loop, stack 8 MiB
  store_code(checked=True) OK · instantiate OK · execute OK (gas 16.101.038.840)   → sin crash

[CONTROL] else implícito pero condición SIEMPRE verdadera (brazo falso nunca tomado)
  store_code OK · instantiate OK · execute OK (gas 16.331.038.840)                 → sin crash

[TRIGGER] else implícito, condición falsa (brazo falso del if = constructo GHSA-v2h8)
  store_code OK · instantiate OK · execute OK (gas 12.881.038.840)                 → sin crash

[TRIGGER] línea recta, 100.000 ifs consecutivos sin target de branch, stack de hilo 256 KiB
  [TRIGGER implicit] execute OK (gas 507.035.390)   [CONTROL explicit] execute OK (gas 668.035.390)
```

Lectura honesta del PoC:

- **Confirmado en local contra el artefacto exacto de producción:** (a) los contratos sintéticos que ejercitan el constructo del advisory **pasan la validación estática de producción** (`store_code(checked=True)` — la misma ruta que `MsgStoreCode`), (b) se compilan y ejecutan en el motor vulnerable, y (c) el brazo falso ejecuta un lowering **distinto** al del control (diferencia de gas 507M vs 668M, y 12.88B vs 16.33B en la variante loop).
- **No reproducí el crash/RCE** con mis shapes sintetizados (el lowering defectuoso requiere layouts concretos de registro/spill que el PoC del reporter —no público— usa; la variante loop además se resincroniza en cada back-edge). El claim crítico queda sostenido por la evidencia de versión: el binario en producción contiene `wasmer-compiler-singlepass 7.4.2`, dentro del rango `>= 7.0.0-alpha.1, <= 7.4.2` del GHSA crítico, con RCE demostrado por el reporter en el mismo stack (CosmWasm/wasmvm, Linux x86-64), y por la regla de oro del encargo (bug en código clonado + ruta activa ⇒ aviso obligatorio aun sin verificación on-chain de TVL).

Trazado de impacto (matemático-operativo, por advisory):

```
Input X  : 1 tx MsgStoreCode (wasm ~3-5 KB que pasa check_wasm) + 1 tx MsgExecuteContract
Estado   : singlepass 7.4.2 compila el `if` de un brazo con block param spill-eado;
           cada toma del brazo falso = +8 bytes de drift de RSP sin cleanup
Output Y : Y > X — el atacante no puso valor alguno (solo gas de 2 txs) y obtiene
           ejecución de código nativo en el proceso del nodo → capacidad de escribir
           estado (minting no autorizado / robo de fondos de terceros) o crash
           persistente del validador. El beneficio neto tras cerrar la secuencia
           es positivo por definición (X = coste de gas ≈ 0 frente a Y = fondos ajenos).
```

---

## 3. HALLAZGOS ADICIONALES (mismo pin, severidad menor / dependiente)

1. **GHSA-xmhf-978x-4pcx (HIGH, 2026-11-02)** — `Table` type mismatch en set/elem/grow/copy/new no verificado; vulnerable **≤ 7.5.0** → afecta el pin (7.4.2) **y también** la línea de remediación 7.5.0; sin parche público («An attacker can potentially place a random address of a function pointer into a funcref Table and later use it as a target or an indirect call»).
2. **GHSA-37h3-8wp9-3hfx (LOW, 2026-11-02)** — `VMContext` embedded `Table` out-of-sync; vulnerable ≤ 7.5.0 → afecta al pin.
3. **GHSA-8v4w-x73c-qp73 (MEDIUM, 2026-11-13)** — singlepass x86_64, expansión SSE de floats incorrecta; vulnerable < 7.6.0, sin parche. **No alcanzable vía CosmWasm** (la validación de determinismo rechaza opcodes float en contratos almacenados) — incluido por completitud del inventario del backend embebido.
4. **`fetchd/Makefile:166-170`** — el comentario dice «The checksums are for … v3.0.8-rc.2» pero los pins (líneas 171-173) son de **rc.3**: verifiqué que los dos sha256 SÍ corresponden a los artefactos rc.3 (x86_64 `ab1a878e…` ✓, aarch64 `cc917523…` ✓). Es documentación obsoleta, no explotable (el build falla cerrado si no coinciden).
5. **Descartados tras verificación** (los fixes ya están en el pin): CWA-2026-001 (mem saturation, parche wasmvm ≥ 3.0.3), CWA-2026-002 (balance ops vía IBCv2, wasmd ≥ 0.61.9), CWA-2026-003 (block stalls, wasmd ≥ 0.61.10), CWA-2026-004 (overflows gas, cosmwasm-vm ≥ 3.0.6), CWA-2026-005 (locals, wasmd ≥ 0.61.14 / wasmvm ≥ 3.0.7 / cosmwasm 3.0.8), CWA-2025-001/002, GHSA-qw47 (atomic.wait32, ≥ 7.4.0... fetchd tiene 7.4.2 ≥ fix notify; wait32 requiere ≥ 7.5.0 — el rango del propio GHSA declara vulnerable solo < 7.4.0), GHSA-r6j4 (atomic.notify, ≥ 7.4.0), GHSA-274f (LLVM, no aplica a singlepass), GHSA-c99m (memory64, no habilitado por cosmwasm-vm), GHSA-6936/5cmw/xprp/p7h3 (WASIX/WASI, no expuesto por cosmwasm-vm).

---

## 4. RESUMEN DE INVARIANTES VERIFICADOS EN EL RESTO DEL CORE

- Integridad de la cadena de build del release: sha256 de ambos artefactos estáticos == Makefile == blobs del fork pinneado == binarios de release v0.15.1-ae12a4f.
- Consistencia fork↔upstream en el pin: cosmwasm-vm 3.0.10-rc.3 == 3.0.10 final (diff vacío en vm/std/crypto); wasmd rc.3 ↔ wasmvm 3.0.8-rc.3 emparejados en el go.mod del propio fork.
- Los binarios rc.3 fueron efectivamente recompilados respecto de rc.2 (blobs distintos) y el x86_64 contiene wasmer 7.4.2/singlepass (strings verificados).
- La ejecución de contratos por el artefacto exacto funciona end-to-end (store→instantiate→execute OK) — la superficie de ataque está viva.

---

## 5. REGISTRO DE AUDITORÍA

**Clones y commits:** ver tabla §1.2. Origen de blobs pesados: tarball `codeload.github.com/fetchai/priv_wasmvm_sec/tar.gz/d89faf1…`.

**Archivos de config/genesis/deploy inspeccionados:** `fetchd/go.mod` (247-279), `fetchd/go.sum` (941-942), `fetchd/Makefile` (105-192), `fetchd/musl.Dockerfile` (100-125), `fetchd/docker-bake.hcl` (43-60), `fetchd/.github/workflows/build_push.yml`, `fetchd/app/app.go` (wiring wasmd 687-708), `fetchd/app/upgrades.go` (42-92), `fetchd/cmd/fetchd/cmd/migrate_genesis.go` (814-830), `fetchd/INTEGRATION.md`, `fetchd/docs/docs/versions.md`, `priv_wasmvm_sec@d89faf1` (`libwasmvm/Cargo.toml`, `Cargo.lock`, `internal/api/bindings.h`, `libwasmvm/src/cache.rs`, `libwasmvm/src/calls.rs`, `libwasmvm/src/memory.rs`, `version_cgo.go`), `priv_wasmd_sec@ab17af8` (`go.mod`), `CosmWasm/cosmwasm@v3.0.10-rc.3/v3.0.10/v3.0.11` (`packages/vm/src/compatibility.rs`, `parsed_wasm.rs`, `calls.rs`, `memory.rs`, `instance.rs`, `wasm_backend/engine.rs`, CHANGELOGs), `CosmWasm/wasmvm` CHANGELOG (v3.0.8/v3.0.9), `CosmWasm/wasmd` CHANGELOG (v0.61.15/v0.61.16), `wasmerio/wasmer` diff `v7.4.2..v7.5.0` (`lib/compiler-singlepass`, `lib/compiler`, `lib/types`), advisories `GHSA-v2h8-j899-r676`, `GHSA-xmhf-978x-4pcx`, `GHSA-37h3-8wp9-3hfx`, `GHSA-8v4w-x73c-qp73`, `GHSA-qw47/r6j4/274f/c99m/6936`, `CWA-2026-001..006`, `CWA-2025-001/002`, releases/tags API de todos los repos citados, `genesis-fetchhub/fetchhub-4/10-software-upgrade-v0.15.1.md`, `fetchhub-4/data/genesis_migrated_5300200.json` (extracción de `app_state.wasm.params`), `dorado-1/` (histórico).

**Tests/PoCs locales ejecutados (comandos y outputs):**

```
xz -dc libwasmvm_muslc.x86_64.a.xz | sha256sum
  → ab1a878ed3beecc5f3821b1fb1f73a5254443e3ba9bca53964a50a122ee77497  (= Makefile:172)
xz -dc libwasmvm_muslc.aarch64.a.xz | sha256sum
  → cc9175235d1e0051002c33a94027a63a008b14dcc13ae483685e511d59754235  (= Makefile:173)

python3 poc.py run explicit_else.wasm            # control A/B
  version_str() = b'3.0.8-rc.3'
  store_code(checked=True) OK · instantiate OK (gas used_int=1035160)
  execute OK (gas used_int=16101038840)                       EXIT=0

python3 poc.py run implicit_else_true.wasm       # control brazo falso no tomado
  execute OK (gas used_int=16331038840)                       EXIT=0

python3 poc.py run implicit_else.wasm            # trigger (else implícito, condición falsa)
  execute OK (gas used_int=12881038840)                       EXIT=0  ← sin crash con este shape

python3 poc.py run-smallstack straight_explicit_else.wasm    # control línea recta, stack 256 KiB
  execute OK (gas used_int=668035390)                         EXIT=0

python3 poc.py run-smallstack straight_implicit_else.wasm    # trigger línea recta
  execute OK (gas used_int=507035390)                         EXIT=0  ← sin crash; lowering distinto confirmado por gas

node -e "new WebAssembly.Module(...)"           # validación estructural de los 5 módulos: VALID
```

Artefactos del PoC conservados en `_audit/poc/` (hashes en `_audit/evidence/POC_ARTIFACTS.sha256`; extracción de dependencias en `_audit/evidence/DEPENDENCY_CHAIN.txt`).

**Repos bloqueados / no accesibles:** `CosmWasm/priv_cosmwasm` (404 — privado; el resto de la org fetchai y de CosmWasm/wasmerio es clonable). Binarios de release no reconstruibles localmente (musl/toolchain del pipeline privado, documentado en §1.3).

**Limitaciones:** sin RPC no se verificó TVL/estado on-chain ni códigos wasm ya almacenados en fetchhub-4; conforme a la regla de oro del encargo, el hallazgo se reporta porque el bug vive en el código clonado (binario pinneado en el rango vulnerable confirmado por el mantenedor) y la ruta de ejecución está activa en producción. La remediación correcta exige actualización coordinada (consensus-breaking) a una línea wasmvm/cosmwasm/wasmer que contenga el fix de GHSA-v2h8 — a la fecha no existe release pública de Wasmer con ese fix; CosmWasm ya publicó la línea v3.0.9/v0.61.16 (wasmer 7.5.0) que fetchd debe adoptar como mínimo.
