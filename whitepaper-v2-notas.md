
## DEFINICIÓN HONESTA DE "PARA TODOS" (reflexión de Edu)
"Para todos" NO = gratis/sin esfuerzo. SÍ = sin barreras de privilegio.
- No necesitas permiso, ni ser rico, ni empresa. No hay club cerrado.
- Con inversión pequeña y al alcance (PC gamer $500-1500 o rentar GPU $/día), cualquiera entra.
- Vs ASICs (decenas de miles), es un salto enorme en accesibilidad.
Analogía: "cualquiera puede tener un changarro" — no gratis, pero accesible sin privilegios.
Los ASICs aparecen si Rupix triunfa (señal de valor, problema de éxito). Para entonces:
comunidad + opción de migrar a Autolykos para proteger el "para todos".
MENSAJE HONESTO: "Para todos: cualquiera con inversión pequeña y al alcance mina, sin
permisos ni privilegios." Realista, justo, sostenible. No prometer "gratis" (mentira/insostenible).

## PÚBLICO REAL DE RUPIX (observación clave de Edu)
Quien mina cripto YA es gente tech/cripto que a menudo YA tiene GPU o la consigue fácil.
"El que sabe de cripto, entiende Rupix y le interesa, sin pensarlo minaría/compraría."
→ GPU (RupixHeavyHash) es PERFECTO para ese público. No hace falta CPU puro (Autolykos).
"Para todos" real = comunidad global cripto/tech (México, India, Argentina...), sin
privilegios, con lo que ya tienen o inversión chica. NO literal "8 mil millones minan".
CONCLUSIÓN FINAL DEL ALGORITMO: RupixHeavyHash (GPU, fácil, 3 archivos) es el camino
correcto — no un compromiso, sino LO ideal para el público real. Autolykos = plan B si
aparecen ASICs (señal de éxito). La pieza más difícil se volvió un plan claro y manejable.

## RESUMEN DEL DÍA (cierre) — todo visible en repo
CERRADO: H-4, H-6 (Kings en proof, el grave), SHA256, textos Kaspa→Rupix, quema explicada en web, pruning mergeado a main.
VIVO: testnet renacida (halving 100k), Coco mina Gold real, PRIMERA TRANSFERENCIA (tx 07c6b1ab...) + quema.
PENDIENTE AUDITOR: cero mentiroso (multi-peer, complejo), H-1 (mapeado), verificable total (commitment header), 22 vulns deps.
IDEAS GUARDADAS: verificador de tx, Muro de Fundadores, contador de nodos, UX wallet gema, blog (Coco), Cerebro de Rupix.
DONDE VAMOS: fin Etapa 2 + corazón de Etapa 3 hecho. ~40% a mainnet. Lo más difícil conceptual (pruning) cruzado.
PARA X (cierre del día): resumen de avances + tx histórica 07c6b1ab...

## ALGORITMO — MAPA COMPLETO (radiografias hechas, listo para cirugia)
ARCHIVOS: pow.go (112) + heavyhash.go (91) + xoshiro.go (38). Chicos, aislados.

EL CORAZON (heavyhash.go):
- Linea 11: matrix [64][64]uint16 (la matriz)
- Linea 15: newxoShiRo256PlusPlus(hash) <- EL PRNG a reemplazar
- Linea 25: computeRank()==64 (NO TOCAR - trampa auditor)

EL PRNG (xoshiro.go) - lo que cambiamos:
- Estado s0,s1,s2,s3 (del hash del bloque)
- Uint64(): res=rotl(s0+s3,23)+s0; t=s1<<17; s2^=s0; s3^=s1; s1^=s2; s0^=s3; s2^=t; s3=rotl(s3,45)
- ESTO tienen los ASICs de Kaspa en silicio.

PLAN CIRUGIA RupixHeavyHash (cambio ESTRUCTURAL del PRNG):
1. Crear rupixprng.go (mismo estado del hash, OTRA formula estructural)
2. heavyhash.go linea 15: newxoShiRo256PlusPlus -> newRupixPRNG
3. Vectores de prueba nuevos (xoshiro_test tiene los viejos)
4. Probar mina/valida/rango64
5. Relanzar testnet (hash distinto, como el commitment)

DECISION PENDIENTE: que formula para RupixPRNG.
- Constante sola = DEBIL (auditor). Cambio ESTRUCTURAL = otro algoritmo/mas operaciones.
- Candidatos: variante con operaciones extra, PCG64, SplitMix, o xoshiro+capa extra.
- Es EL corazon de la seguridad -> pensar con calma.

TRAMPAS AUDITOR: (1) rango 64 (2) no desbordar uint16 (3) no tocar computeRank (4) vectores nuevos.

## BUG CRITICO: LA FORJA NO ENTRA AL MEMPOOL (diagnostico 11-sep, cazado por el commitment)

SINTOMA: el wallet dice "Ascenso forjado — gema Diamante creada" con TxID,
pero la gema NO existe en la cadena (gemsCommitment sigue 2f71eee = 0 gemas).

DIAGNOSTICO (confirmado):
1. La forja se ARMA bien: forge_internal.go crea output de gema (version Diamante,
   linea 68) + output de quema OpReturn 0x6a de 10 Gold (linea 69). Correcto.
2. El wallet marca el UTXO como usado LOCALMENTE (broadcast.go: s.usedOutpoints).
   Por eso la 2da forja dice "already spent in the memory pool" — es el bloqueo
   LOCAL del wallet, no el mempool real.
3. La tx NUNCA entra al mempool del nodo (GetMempoolEntries = 0 en todo momento,
   monitoreado t=2,4,6,8,10s).
4. El nodo NO da error visible (rpcclient SI revisa response.Error y esta bien;
   el handler HandleSubmitTransaction pone el error en la respuesta).
5. La red MINA normal (bloques suben). checkLevelRules se ve correcto (auditor lo
   valido con 400k casos). El commitment CUENTA bien (calculateGemsHistory por
   version de output).

EL PUNTO EXACTO A CAZAR (fresco):
- La tx de forja se pierde entre "wallet la envia" y "mempool la registra".
- Sospechoso #1: el output de gema con ScriptPublicKey.Version=1 (Diamante).
  La validacion estandar del mempool (transaction_in_isolation / mass / script)
  puede rechazar outputs con version != 0 ANTES de checkLevelRules, silenciosamente
  o con un error que no se propaga.
- Sospechoso #2: DomainTransactionToRPCTransaction / RPCTransactionToDomainTransaction
  puede no serializar bien el output con version de gema (se pierde el version en el
  viaje wallet->RPC->nodo).
- Sospechoso #3: la conversion del output de gema en el mempool.

PLAN DE ARREGLO (fresco):
1. Poner un log temporal en mempool.validateAndInsertTransaction para ver si la
   forja llega y que error da.
2. O revisar DomainTransactionToRPCTransaction: serializa el ScriptPublicKey.Version?
3. Verificar que la validacion estandar del mempool acepte outputs version 1-4 (gemas).

IMPORTANTE: el HITO 2 (Coco forja) NO esta completo. La wallet de Coco muestra
"1 Diamante" pero es LOCAL (usedOutpoints/conteo por UTXO version), NO esta en la
cadena. El commitment lo destapo: 0 gemas reales. Sin el commitment, habriamos
celebrado un Diamante falso. El commitment hizo EXACTAMENTE su trabajo.

BUG SECUNDARIO: el wallet dice "creada" aunque la tx no se confirme. Deberia
esperar/verificar que entre al mempool antes de reportar exito.

## BUG FORJA - DIAGNOSTICO PROFUNDO (11-sep, sesion larga)

CONFIRMADO con logs temporales en todo el camino:
1. Wallet arma la forja bien (2 inputs, 3 outputs: gema+quema+cambio)
2. [SUBMIT-DEBUG] la tx LLEGA al handler HandleSubmitTransaction
3. [FLOW-DEBUG] AddTransaction llama a ValidateAndInsertTransaction
4. [FLOW-DEBUG] dice "ACEPTADA en mempool" -- PERO tarda ~3 SEGUNDOS raros
   (de 06:07:26.336 a 06:07:29.424) y coincide con un "Accepted block"
5. El mempool SIEMPRE muestra 0 tx (monitoreado 20 veces cada 0.5s: nunca aparece)
6. El commitment SIGUE en 2f71eee (0 gemas) tras decenas de forjas
7. Los bloques se minan CON otras tx normales (8 tx en 4 bloques) pero NO la forja

DESCARTADO:
- checkLevelRules (auditor lo valido, codigo correcto)
- MaxScriptPublicKeyVersion (=4, permite gemas 1-4)
- GetScriptClass (solo mira bytes, gema es script estandar)
- El conteo del commitment (calculateGemsHistory correcto)
- La validacion del mempool (los logs FORJA-DEBUG nunca dispararon = no llega a rechazarse ahi)

EL MISTERIO: la forja se "acepta" (FLOW-DEBUG ACEPTADA) pero:
- No aparece en el mempool (0 siempre)
- Tarda 3 seg raros en "aceptarse"
- No se mina (commitment sigue 0)

SOSPECHA PRINCIPAL (rematar fresco): el bug esta entre "ACEPTADA en mempool" y
"el minero la incluye". Revisar:
1. Como el blockTemplateBuilder selecciona tx del mempool (quiza excluye las de
   forja por el output version != 0, o por masa, o por el OpReturn de quema)
2. Por que tarda 3 seg (PopulateMass? ValidateTransactionAndPopulate con la gema?)
3. Si la tx se acepta pero se remueve inmediato por revalidacion/conflicto UTXO
4. El transactionsPool.addTransaction: la agrega pero GetMempoolEntries no la ve?
   (quiza se agrega a un pool que el minero no consulta, o hay dos pools)

ENFOQUE FRESCO: leer blocktemplatebuilder + como el minero pide tx del mempool +
transactionsPool.addTransaction. El bug esta en la seleccion de tx para el bloque,
no en la validacion (que pasa) ni en el conteo (que funciona).

LOGS TEMPORALES PUESTOS (quitar al arreglar):
- validate_and_insert_transaction.go: RUPIX-FORJA-DEBUG (3 logs)
- submit_transaction.go: RUPIX-SUBMIT-DEBUG (2 logs)
- flowcontext/transactions.go: RUPIX-FLOW-DEBUG (3 logs)

## BUG FORJA - CAUSA RAIZ FINAL ENCONTRADA (12-sep) ✅

EL BUG (100% confirmado con log RUPIX-COMMIT-DEBUG):
Al minar un bloque con una forja, el header lleva gemsCommitment de "0 gemas"
pero la validacion calcula "1 Diamante" -> MISMATCH -> bloque rechazado -> timeout.

LOG DE LA PRUEBA:
MISMATCH bloque fc752026...: header=2f71eee (0 gemas) calculado=780e9027 (D=1)

CAUSA EXACTA:
block_builder.go newBlockGemsCommitment() (linea 332) calcula el sello usando
el gemsHistory del VIRTUAL (model.VirtualBlockHash) = estado ACTUAL = 0 gemas.
NO incluye las forjas de las tx que el bloque va a incluir.

Pero verify_and_build_utxo.go (linea 41) valida con calculateGemsHistory(blockHash,
acceptanceData) que SI cuenta las forjas del bloque -> 1 Diamante.

Template: 0 gemas. Validacion: 1 Diamante. Mismatch -> rechazo.

FIXES YA APLICADOS HOY:
1. BlockCandidateTransactions (mempool.go): exime forjas del anti-spam (esForja).
   Las forjas ya no se filtran como spam. APLICADO Y COMPILA.

FIX PENDIENTE (el final):
newBlockGemsCommitment debe calcular el gemsHistory INCLUYENDO las forjas de las
tx del bloque, igual que la validacion. Opciones:
- A) Recalcular gemsHistory sumando las gemas creadas/quemadas en selectedTxs
- B) Usar el mismo calculateGemsHistory con el bloque candidato antes de sellar
- Reto: en el momento de armar el template, las tx aun no tienen acceptanceData.

OTRO HALLAZGO: el minero da timeouts de 10s al submitir bloques que se rechazan
(reintenta), por eso la red minaba lento durante las pruebas.

LOGS TEMPORALES A QUITAR: RUPIX-COMMIT-DEBUG en verify_and_build_utxo.go (el resto
ya se quitaron: SUBMIT, FLOW, FORJA, READY).

## BUG FORJA - AVANCE 12-sep noche (fix anti-spam OK, falta alinear gemsHistory virtual/padre)

LOGRADO HOY:
1. Fix anti-spam en BlockCandidateTransactions (exime forjas). APLICADO.
2. Fix en newBlockGemsCommitment: recibe transactions y cuenta forjas del bloque
   (Clone + nacidosNetos por nivel). APLICADO Y COMPILA.
3. La forja se arma, se acepta en mempool, se mina. gems del wallet = 1 Diamante.
4. Desbloqueo temporal para pruebas: levels.go return uint64(level)*10 (QUITAR despues).

EL BUG RESTANTE (sutil, de arquitectura):
- Template (block_builder.newBlockGemsCommitment) lee gemsHistory del VIRTUAL
  (model.VirtualBlockHash) -> da 2f71eee (0 gemas).
- Validacion (verify_and_build_utxo.calculateGemsHistory) lee del PADRE del bloque
  (SelectedParent) + cuenta forjas -> da 780e9027 (1 Diamante).
- Los dos NO coinciden -> MISMATCH en cada bloque -> se rechaza -> red trabada.

LA CAUSA PROFUNDA:
El gemsHistory del VIRTUAL no se actualiza/lee igual que el de los bloques minados.
El fix de contar forjas del bloque candidato NO basta porque el problema es que el
virtual y el padre dan bases distintas.

PARA REMATAR FRESCO:
- Entender como se actualiza gemsHistoryStore para el VIRTUAL vs bloques normales.
- Quiza el template debe leer del SelectedParent del virtual (no del virtual mismo),
  igual que la validacion lee del SelectedParent del bloque.
- O asegurar que el gemsHistory del virtual incluya las forjas ya confirmadas.

ESTADO: forja se mina (gems=1) pero template y validacion sellan distinto.
El commitment funciona (detecta el mismatch). Falta alinear los dos calculos.

LOGS TEMPORALES ACTIVOS: RUPIX-COMMIT-DEBUG en verify_and_build_utxo.go
DESBLOQUEO TEMPORAL: levels.go *10 (QUITAR, volver a *blocksPerHalving)

## BUG FORJA - DIAGNOSTICO FINAL COMPLETO (12-sep, 2 dias de cirugia)

3 BUGS CAZADOS Y ARREGLADOS:
1. Anti-spam (BlockCandidateTransactions): exime forjas. OK.
2. Wallet coinbaseMaturity: era uint64(1000) hardcodeado en server.go:100,
   cambiado a 100 (como el nodo). AHORA EL GOLD SE GASTA. OK.
3. newBlockGemsCommitment: recibe transactions y cuenta forjas del bloque. OK.
4. verify_and_build_utxo: agrega Stage del gemsHistory/kingsCount tras validar. OK.

EL BUG RAIZ QUE FALTA (arquitectura del gemsHistory del VIRTUAL):
- El template (newBlockGemsCommitment) lee gemsHistory del model.VirtualBlockHash.
- El VIRTUAL NO acumula las forjas ya confirmadas: sigue en 0 gemas.
- La validacion lee del PADRE del bloque (que SI tiene el Diamante de la forja anterior).
- Bloque nuevo: template=2f71eee (virtual 0 gemas), validacion=780e9027 (padre 1 Diamante).
- MISMATCH persiste -> el bloque se rechaza.

LA FORJA YA SE EJECUTA (Gold se gasta, crea el Diamante, MISMATCH muestra D=1).
El problema es SOLO que el gemsHistory del VIRTUAL no refleja las forjas confirmadas.

PARA REMATAR FRESCO:
- Entender como se actualiza el gemsHistory del VIRTUAL tras aceptar un bloque.
- El VIRTUAL debe recalcular su gemsHistory incluyendo las forjas confirmadas,
  igual que el UTXO del virtual se actualiza.
- Buscar donde se hace el "resolveVirtual" o "updateVirtual" y agregar el gemsHistory.
- Quiza el Stage debe hacerse tambien para VirtualBlockHash, no solo por blockHash.

DESBLOQUEO TEMPORAL ACTIVO: levels.go *10 (QUITAR).
LOG TEMPORAL: RUPIX-COMMIT-DEBUG en verify_and_build_utxo.go.

## HITO 2 COMPLETO - DIAMANTE REAL EN LA CADENA (14-sep)
- Testnet cruzo 100,000 bloques -> Diamante desbloqueado
- Diamante REAL forjado: gems=1, commitment=780e9027 (sellado en cadena)
- SIN mismatch, red fluida, codigo AUDITADO (H-8, H-9, H-10 cerrados)
- Test del King en el CI (regresion H-10 cazada automatico)
- Binario fresco 00:47 con todos los fixes
- El verificable total, probado en vivo en la red real.
- De "la forja no funciona" (viernes) a "Diamante real sellado" (domingo).

## DEUDA DE TESTS - diagnostico honesto (14-sep)
PRODUCCION FUNCIONA (go build ./... limpio, Diamante real sellado en vivo).
Tests del commitment/escalera PASAN (corpus 9388/0, King, blockbuilder).

DEUDA: 26 paquetes de test fallan, DOS tipos:
- TIPO A (5, "build failed"): format string que Go endurecio. HEREDADO de
  Kaspa (logs.go:183, standard_test.go:282). El codigo compila, el test no.
- TIPO B (21, panic nil): helpers de test viejos crean headers con hashes nil;
  al serializar a disco petan. Mezcla heredado + gems. Guards agregados en
  consensushashing/block.go y serialization/blockheader.go ayudan pero no curan
  todo (hay mas caminos con nil).

PENDIENTE (sesion fresca): limpiar TIPO A (mecanico) + TIPO B (cuidadoso).
Meta: go test ./... 100% verde ANTES de sellar "todo verde".
NO se sella todo-verde hasta limpiar. Se documenta la verdad.

## LIMPIEZA DE TESTS - avance 14-sep (26->18 paquetes)
LO CURADO HOY:
- TIPO A (5 format strings heredados de Kaspa): Wrapf->Wrap, Fprintf->Fprint,
  Errorf->New, Criticalf con %s, %q->%v. 5 build-failed corregidos.
- Framework de test header 1 (test_block_builder.go:105, buildBlock):
  gemsCommitment de ceros/nil -> GenesisGemsCommitment() (sello de 0 gemas).
  Curo dagtopologymanager + 3 mas.
- Guards contra gemsCommitment nil (consensushashing, serialization).

RESULTADO: 26 -> 18 paquetes FAIL. OK subio a 43+.

PENDIENTE FINAL (antes de mainnet) - EL HEADER 2:
- test_block_builder.go:138 (buildHeaderWithParents) tambien usa
  GenesisGemsCommitment() FIJO (0 gemas). Pero ese camino RECONSTRUYE
  bloques con historia -> el gems real puede ser != 0.
- La validacion calcula el gems real (ej f7dfa586) y el header sella 0 ->
  mismatch -> DisqualifiedFromChain en ~18 tests de consensus.
- FIX: buildHeaderWithParents debe CALCULAR el gems real (como produccion),
  no ponerlo fijo. Tiene los datos: tempBlockHash, ghostdagDataStore,
  acceptanceData, stagingArea. Falta acceso a calculateGemsHistory (metodo
  privado de consensusStateManager) - exponerlo o replicar la logica.
- DIAGNOSTICO 100% hecho. Solo falta aplicar el calculo. Sesion fresca.
- NO afecta produccion (testnet real funciona, Diamante sellado). Es deuda
  de tests del framework de consensus.

METODO CLAVE APRENDIDO: log.Warnf se TRAGA en tests; fmt.Printf SI sale.
Por eso el diagnostico tardo - los logs no aparecian.

## HITO HISTORICO - COMUNIDAD REAL (14-sep, 23:20)
PRIMER FORJADOR EXTERNO: Coco forjo 3 Diamantes REALES desde su propio
nodo/wallet/computadora. Validacion TOTAL con comunidad:
- Edu envio 50 RUPIX a Coco (transferencia entre personas) - llegaron
- Coco forjo 3 Diamantes (tx adbfe36b...) desde su nodo externo
- El Gold se quemo (Coco confirmo: "si quemo, si se movio el balance")
- Las forjas se minaron (mempool 0, en la cadena)
- El commitment de la RED cambio a 6c8f913d (refleja las gemas de Coco)
- Red fluida (203k+ bloques, sin trabarse)
TOTAL en la red: 4 Diamantes (1 Edu + 3 Coco), sellados en el commitment.
"Parece que todo al 100 mi chilps" - Coco.
El verificable total funcionando entre DOS personas reales.
We are all Rupix - ya no es lema, es hecho.
