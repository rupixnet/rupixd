
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
