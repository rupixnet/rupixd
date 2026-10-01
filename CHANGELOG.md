# Changelog — Rupix

Historial honesto de cambios. Todo verificable contra los commits del repositorio.

Formato: [versión] - fecha - descripción técnica

---

## [Sin publicar]
- **CAMBIO DE CONSENSO (rama `tope-no-acepta`, `826be325`, para v0.6.2; no corre en el seed):** el tope histórico de gemas se aplica al aceptar la transacción, no al insertar el bloque. Antes, un bloque con una forja sobre el tope se rechazaba entero, y un atacante podía "envenenar" la punta: los bloques honestos que lo mergearan caían con él (hallazgo del auditor, 30-sep). Ahora el bloque entra y la forja que excede el tope no se acepta (`applyMergeSetBlocks` → `maybeAcceptTransaction`, conteo del selected parent en orden GHOSTDAG, `utils/topes`), sin efectos colaterales: el Gold sigue vivo, el sello no cambia, los hermanos honestos son UTXOValid. El mempool aplica el mismo conteo en la puerta (`ErrGemsCapExceeded`). Tests en rojo antes y en verde después: `TestVenenoDeTopeNoMataAlHonesto`, `TestDosHermanosUltimaGema` (dos consensos coinciden), `TestForjaSobreTopeEnMempool`; suite `./domain/...` en verde (35 paquetes).
- Wallet (rama `wallet-principiantes`, 3/3): `create` pide dos palabras al azar de la frase semilla para comprobar que quedo en papel (ninguna bandera la salta; `--yes` solo sobrescribe el archivo); si el daemon no esta corriendo, cualquier comando lo dice en palabras y da el comando exacto para arrancarlo con la misma red. `LICENSE`: se agrega la linea de copyright de Rupix (2026) conservando las de Kaspa, btcsuite, Decred y Conformal.
- Wallet en espanol o ingles (rama `wallet-principiantes`, para v0.6.2): `idioma.go` con tabla clave -> {es, en}; `--lang`, `RUPIX_LANG`, comando `language es|en` (guarda la preferencia junto a las llaves), idioma del sistema, y espanol por defecto. Traducido lo que ve un principiante: crear (frase semilla), saldo, gemas, forjar, enviar, transferir gema, direccion nueva, confirmaciones, contrasena y errores explicados. Test: cada clave con las dos versiones y los mismos %s/%d. `balance` dice "Tienes X RUPIX (+ Y llegando)". Guias de forja ES/EN con la seccion "Que cambio y por que".
- Wallet para principiantes (rama `wallet-principiantes`, para v0.6.2): `forge` revisa antes de mandar (gemas y Gold contra lo que pide la escalera, y si el nivel esta abierto preguntando al nodo), resume lo que va a pasar y pide confirmacion; `send` y `transfer-gem` confirman antes de pedir la clave; `--gem-address` opcional (primera direccion); rechazos conocidos del nodo explicados en palabras sin esconder el original; `--yes` para scripts. Probado en vivo contra el daemon del seed sin mover nada.
- `tools/verificar-binarios.sh`: recompila un tag con los flags del CI y compara SHA256 con la release. v0.6.1: 4 de 4 identicos (compilacion reproducible verificada).
- CI: `race.yaml` heredado (corria sobre `master` y ramas `v*-dev` inexistentes) retirado; el nuevo corre `go test -race ./...` sobre `main` cada noche. Badge del CI en los README.
- CI: el workflow `Tests` heredado (rojo en cada push por piezas de Kaspa) se retira a `.github/workflows-retirados/` y se reemplaza por gofmt + go vet + staticcheck + build + `go test ./...` en Linux y macOS.
- Wallet: el output de quema entra al mock de `estimateFee` antes de armarlo (antes se agregaba despues y no contaba; lo tapaba el colchon de 100 rupias). Encontrado por staticcheck SA4006.
- Test del King: `disp` sin uso (staticcheck SA4006).
- Pendiente (v0.6.2): validar la lista `MsgPruningPoints` contra los checkpoints despues de `ArePruningPointsInValidChain` (proteccion para nodos que sincronizan desde cero). Hasta entonces el checkpoint protege a nodos ya sincronizados.
- Pendiente: RPC con los conteos de gemas (`GemsHistory` del virtual) y contador exacto de Gold minado/quemado, fuera del sello; tarjetas en explorador y web.

## [v0.6.1] - 2026-09-29 — Checkpoints para DAG, wallet usable, primer checkpoint publicado
Cuatro ramas revisadas y aprobadas por el auditor externo el 29-sep, mergeadas sin conflictos sobre `1e38bace`. Sin cambios de consenso salvo la regla de checkpoints (que hasta hoy no tenia efecto: la lista estaba vacia) y el primer checkpoint de la testnet.
- **Primer checkpoint publicado (testnet #4):** blue score 86,400, hash `7e2ece393c7d991c86e7ba915276cd85b5fc19e8647d5d197fa26bf116604fad` (DAA 86,399), caduca en DAA 2,000,000. Es el punto de poda que el seed y el nodo de JC ya compartian. `TestCheckpointsPublicados` obliga a que el codigo y CHECKPOINTS.md digan lo mismo.
- Emision: `TestTotalSupply` afirma la emision por calendario (41,999,994.96) y documenta el exceso acotado en cada frontera de halving (medido: 0.5 RUPIX en el halving 2). `MaxRupia` es tope por transaccion, no de emision. Ver MEMORIA.
- Guias ES/EN: la forja ya no lleva `--password` en la linea; la frase semilla se anota en papel al crear la wallet.
- Envio grande probado en la red publica (29-sep, wallet de la rama `wallet-envios-grandes`, daemon de prueba): 1,000 RUPIX en 24 transacciones / 8.6 s; 10,000 RUPIX en 232 transacciones / 46.1 s, desde una wallet de minero de ~78k pedazos. Antes 1,000 se cortaba a los 120 s.
- Arreglo del envio grande (rama `wallet-envios-grandes`, para v0.6.1): la seleccion de pedazos estima la comision en tiempo lineal y la exacta se calcula una sola vez al final. Medido en el seed, misma wallet de 78k pedazos: 100 RUPIX 1.6s -> 0.10s; 300: 16s -> 0.24s; 600: 65s -> 0.47s; 1,000: cortado a 120s -> 0.82s.
- Checkpoints para DAG (rama `checkpoint-dag`, para v0.6.1): la regla pasa a "todo bloque con blue score >= X + MergeDepth debe tener al bloque canonico H en su pasado". Un hermano tardio de H ya no rompe nada; una historia alterna sin H se rechaza. `TestCheckpointDAG` lo prueba, y se verifico que la prueba FALLA si la regla se apaga.
- Mempool (rama `mempool-niveles`, para v0.6.1): prueba automatica de que una forja de un nivel cerrado se rechaza al entrar al mempool y la misma transaccion entra pasado su halving (H-1 nivel A).
- `.gitignore`: los binarios quedan anclados a la raiz; la linea `rupixwallet` tapaba la carpeta `cmd/rupixwallet` y hacia fallar `git add` ahi.
- `tools/verificar-emision.py`: compara el Gold que existe en tu nodo con la regla de emision (enteros exactos) y alerta si alguna vez hay mas del permitido.
- Wallet (encontrado el 28-sep, se arregla en v0.6.1): enviar mucho Gold desde una wallet de minero se corta a los 2 minutos porque la seleccion de pedazos recalcula la comision rearmando toda la transaccion (costo al cuadrado; medido en el seed). `parse` imprime "KAS".
- Web: pagina `matematica.html` (emision, escalera, costo de mover, calculadora) en espanol e ingles.
- `gofmt` en todo el repo (251 archivos): solo formato. Los unicos cambios que no son espacios son imports reordenados (el renombre kaspa->rupix cambio su orden alfabetico). Compila igual.
- Explorer: la copia de la pagina en el repo estaba vieja (halving de testnet en 10,000; el real es 100,000). Sincronizada con la viva, y el tiempo al proximo halving ahora se muestra en minutos, horas o dias.
- `govulncheck` (27-sep): 0 vulnerabilidades que afecten al codigo; 3 en modulos requeridos que el codigo no llama.
- `SECURITY.md`: como reportar una vulnerabilidad en privado (pestana Security de GitHub), con tiempos de respuesta.
- Builds de release con `-trimpath -buildvcs=false -buildid=`: primer paso hacia builds reproducibles. Falta comprobar que dos builds den el mismo hash.
- `params.go`: el comentario del halving de mainnet decia "~decadas"; son ~16 meses (42M bloques a 1/seg). Hallazgo del auditor. Al revisarlo aparecio `constants.BlocksPerHalving = 150`, una constante sin uso cuyo comentario decia que definia la emision: eliminada. El consenso usa el valor de cada red.
- Alarma v2.1: sigue la rotacion del log del nodo (guarda inode + posicion; si el archivo viejo no aparece, ERROR). Pedido del auditor al revisar la v2.
- `go.mod`: minimo de Go de 1.26.6 a 1.25.0 (la minima real del arbol de dependencias). Un revisor con Go 1.25 ya puede compilar y correr la suite. Los binarios de release se siguen compilando con 1.26.6.
- Residuos del fork: el daemon de rupixwallet escribe su log en `~/.rupixwallet` (antes `~/.kaspawallet`); el minero dice "Rupixd is not synced" (antes "Kaspad").
- Alarma de bloques descalificados **corregida**. La v1 (25-sep, pedida por el auditor el 23-sep) caminaba hacia atras por la cadena seleccionada y, por construccion, nunca podia ver un bloque descalificado: dijo "OK" 134 veces sin poder sonar. La v2 lee la decision del propio nodo en su log (solo lo nuevo desde la revision anterior) y dice ERROR si el log no crece. Probada con un log falso: suena. Falta probarla provocando un bloque descalificado en devnet.
- Nodo: `resolve_block_status.go` registra cada bloque descalificado en nivel Warn (`BLOQUE DESCALIFICADO <hash>: <motivo>`); antes era Debug y no se veia. Cambio de log, no de consenso: no requiere relanzamiento.
- README (es/en): seccion "Como minar Rupix".
- `changelog.txt` renombrado a `CHANGELOG-kaspad-upstream.txt` para no confundirlo con este archivo.

## [v0.6.0] - 2026-09-26 — Tres cambios de consenso + testnet #4
- Dominio keccak propio: `heavyHashDomain` pasa de "HeavyHash" a "RupixHeavyHash". El hash de Rupix es distinto al de Kaspa desde el primer bit.
- `computeRank` con aritmetica entera (mod 2^61-1) en lugar de float64 con tolerancia: el resultado es identico en cualquier CPU y sistema operativo. Era el riesgo silencioso mas grave segun el auditor; el lo verifico por su cuenta.
- Fix del bug del King: `newBlockGemsCommitment` sumaba las forjas del propio bloque ademas de las del mergeset, asi que el minero sellaba un conteo distinto al que calcula el validador. El primer King real habria sido rechazado. Lo cazo `TestKingsEndToEnd`, que falla si se revierte el fix.
- Coinbase con `Version: 0` en lugar de `MaxScriptPublicKeyVersion`.
- Testnet relanzada (#4) con los tres cambios. Vectores de `pruning_test` regenerados.

## [v0.5.2] - 2026-09-20 — Suite de tests en verde
- `go test ./...` de 18 paquetes rojos a 0, sin tocar consenso. Causa raíz (predicha por el auditor como H-5): tests heredados creaban outputs con `Version = MaxScriptPublicKeyVersion` (= 4 = Kings en Rupix) y la escalera los rechazaba. Corregido en el framework de test, vectores bip32/txscript/ventana/poda regenerados desde el código. `DisasmString` aceptaba solo la versión máxima; ahora 0..4. `go vet` limpio en producción.
- README en inglés + aclaración: fork de kaspad, cadena independiente, no usa KAS.

## [v0.5.1] - 2026-09-18 — Checkpoints temporales
- Defensa contra el 51% mientras el hashrate es bajo: bloque en el DAA score de un checkpoint debe tener el hash canónico o `ErrCheckpointMismatch`. Caducidad dentro del consenso. Probado en devnet (correcto acepta, falso rechaza). Lista vacía: compatible con v0.5.0, sin relanzamiento. Política en CHECKPOINTS.md.
- Primer Diamante con RupixHeavyHash en la red pública (commitment `780e9027…`). README bilingüe.

## [v0.5.0] - 2026-09-16 — RupixHeavyHash (algoritmo propio)
- Generador de la matriz de kHeavyHash reemplazado (xoshiro256++ → rupixPRNG con multiplicación y sello RUPIX). Los ASIC de Kaspa quedan fuera. Honesto: ventaja de meses, no independencia. Génesis re-minado (misma fecha y mensaje). Testnet relanzada (#3).
- Versión del binario corregida (reportaba 0.4.0). Dependencias: grpc 1.83.2, 0 vulnerabilidades reales. Guard del bucle de generateMatrix.

## [v0.4.5] - 2026-09-14 — Diamante real, primer forjador externo
- Bug de la forja resuelto (el virtual no calculaba su gemsHistory). Primer Diamante real sellado (`780e9027…`). Segundo usuario forjó 3 Diamantes desde su nodo. Primer Platino. Auditoría externa: H-8, H-9, H-10 cerrados.
- **v0.4.4 marcada como defectuosa — no usar** (regresión de verificación introducida por un fix).

## [v0.4.2] - 2026-09-12 — Commitment en header (verificación total)
- El conteo de gemas se sella en el hash de cada bloque, protegido por PoW, validado al recibir, persistido en disco. Halving de testnet a 100,000 bloques. Testnet relanzada (#2). Pruning verificable (gemsHistory viaja en el proof).

## [v0.4.0] - 2026-09-03 — Testnet pública
- Primera testnet pública con binarios verificables (SHA256, compilados por CI). Génesis sin premine (04/03/2026 — RUPIX IS ALIVE). Escalera completa en consenso (5 niveles, 10:1, techos históricos). Burn por transacción. Explorador público. Primer nodo externo.

## [Reinicio] - 2026-06-13 — Desde Kaspa limpio
- El código anterior (v0.2.x, con 1,700 líneas de una IA previa) se descartó por bugs estructurales en cascada. Rupix se reconstruyó desde kaspad limpio, con la economía como ley. Todo lo anterior a esta fecha es historia, no código vigente.

## [v0.2.2] - 2026-05-25 — Primera testnet sincronizable

### Milestone histórico

Por primera vez en la historia del proyecto, un nodo Windows con base de
datos limpia logra sincronizar completamente desde el nodo semilla de
Hetzner y registra en log:

    [INF] PROT: IBD with peer ... finished successfully

Hora exacta: 25 mayo 2026, 23:46 hora México (05:46 UTC).

### Fixed

- **FIX-002 (FinalityDuration testnet)**: cambiado de `1 * time.Minute` a
  `12 * time.Hour`. El valor anterior violaba la invariante de Kaspa
  `MergeSetSizeLimit << FinalityDuration/TargetTimePerBlock`, causando
  poda agresiva e `ErrPrunedBlock` durante IBD. El valor `12 * time.Hour`
  es el estándar de Kaspa testnet.

- **FIX-003 (MaxBlockParents testnet)**: cambiado de `1` a `10`. Con
  `MaxBlockParents=1` el DAG no convergía: una red de 695 bloques
  generaba 196 tips paralelos (28% sin merge), causando
  `ErrMissingParents` durante header sync porque los headers llegaban
  fuera de orden topológico. Valor `10` es el estándar de Kaspa y es
  coherente con `defaultGHOSTDAGK=18`.

### Scope

Solo se modificaron parámetros de **testnet**. Mainnet, Simnet y Devnet
sin cambios.

### Conocido (pendiente)

- `ErrMissingParents` aparece marcado como "non-critical" durante relay
  post-IBD. El nodo lo maneja correctamente, pero se debe investigar si
  el comportamiento es óptimo o si hay margen de mejora.
- Pendiente validación: cliente Windows debe alcanzar `blockCount` igual
  al seed y exponer balance correcto vía `GetBalanceByAddress`.

### Verificable

- Commit del fix: `a01809d3`
- Commit del merge: `b83e9c41`
- Tag: `v0.2.2`
- Log de prueba: cliente Windows con BD limpia (`C:\\RUPIX\\ibd-test-fix003.log`)

---

## [v0.2.1] - 2026-05-21 — IBD-001 cerrado

### Fixed
- **IBD-001 (crítico)**: nodo Windows con base de datos limpia ya no crashea
  con `nil pointer dereference` al sincronizar desde cero.
- Causa raíz: la función `ValidateAndInsertBlockAsTrusted` saltaba las
  validaciones que construyen y almacenan GHOSTDAG data. Sin esa data,
  llamadas subsiguientes al store retornaban not-found, lo cual era
  enmascarado por parches que devolvían datos vacíos con `selectedParent=nil`,
  causando el nil pointer en el pruning manager.
- Fix aplicado: alineamos el flujo IBD con el comportamiento de kaspad
  upstream (referencia: kaspad master, `ibd.go` líneas 519 y 691).

### Removed
- Atajo inseguro "ignore not found durante IBD" en `processHeader` 
  (silenciaba errores legítimos).
- Import huérfano `reachabilitydata` en `reachability_data_store.go`.

### Known issues exposed
- **ErrPrunedBlock**: con el nil pointer eliminado, sale a la luz que
  el testnet poda demasiado agresivo por `FinalityDuration=1min`.
  Pendiente para FIX-002.

### Verificable
- Commit del fix: `00d2f969`
- Commit de merge: `e8a006c7`
- Tag: `v0.2.1`

---

## Cómo verificar el código

```bash
git clone https://github.com/rupixnet/rupixd.git
cd rupixd
git checkout v0.2.1
go build -o rupixd .
```

No confíes, verifica.

