# CIRUGÍA DEL PRUNING VERIFICABLE — mapa (rama pruning-verificable)

## EL PROBLEMA
El pruning manager (pruningmanager.go, 1264 líneas) NO menciona las gemas.
El conteo de gemas (gemshistorystore) está DESCONECTADO del pruning point.
Un nodo nuevo recibe el estado de monedas pero NO una prueba del conteo de gemas
→ tiene que confiar (rompe "verifica") o bajar todo (rompe "para todos").

## LO QUE JUEGA A FAVOR
- El conteo es función PURA de la cadena: conteo(bloque)=conteo(padre)+nacidos.
  Encadenado, inmune a reorg (confirmado en auditoría ronda 2).
- Kaspa YA tiene pruningproofmanager (prueba de pruning al sincronizar).
- gemshistorystore YA tiene: Stage, Get, Delete, serializeHistory.

## LA CIRUGÍA (3 pasos)
1. Que el gemshistory del pruning point VIAJE al nodo nuevo (ya se serializa).
2. Que el nodo nuevo lo VERIFIQUE (recalcule contra el pruning proof) — cuadra o rechaza.
3. De ahí en adelante, el nodo calcula normal (ya funciona).

## FUNCIONES CLAVE
- pruningmanager.go:115 UpdatePruningPointByVirtual (decide/guarda el punto)
- pruningmanager.go:185 savePruningPoint (sella el punto)
- consensusstatemanager/gemshistory.go:23 calculateGemsHistory
- gemshistorystore/gemshistory_store.go: Stage/Get/serializeHistory

## DIFICULTAD: MEDIA (bajó de ALTA al entender el terreno)
No requiere criptografía nueva. Es conectar y validar piezas existentes.
PERO toca el pruning proof (zona delicada) → rama, tests, devnet antes de main.

## SIGUIENTE PASO (próxima sesión)
Estudiar pruningproofmanager a fondo para saber EXACTAMENTE dónde enganchar
el gemshistory. Luego implementar en pequeño, con tests, en esta rama.

## MAPA QUIRÚRGICO COMPLETO (radiografía 4 - pruningproofmanager)
La PruningPointProof es la "caja" que viaja al nodo nuevo. Agregar 1 campo (GemsHistory) y engancharlo en 3 puntos:

- PUNTO 1: BuildPruningPointProof (pruningproofmanager.go:104)
  → agregar el gemshistory del pruning point a la caja (usar gemshistorystore.Get + serializeHistory)
- PUNTO 2: ValidatePruningPointProof (pruningproofmanager.go:318)
  → recalcular el conteo desde los headers de la caja y confirmar que cuadra. Si no: RECHAZAR. (ESTO es el "verificable")
- PUNTO 3: ApplyPruningPointProof (pruningproofmanager.go:837)
  → guardar el gemshistory verificado.
- + Ampliar el tipo PruningPointProof con campo GemsHistory.

## DIFICULTAD FINAL: MEDIA (sin criptografía nueva)
1 estructura + 3 funciones + store existente. Con tests en cada punto.
DELICADO: es el pruning proof. Error = nodos rechazan/aceptan mal. Rama + tests + devnet.

## PRÓXIMA SESIÓN: empezar a implementar, empezando por ampliar el tipo PruningPointProof, luego PUNTO 1 (build), con tests.

## AVANCE SESIÓN (rama pruning-verificable) — FLUJO INTERNO COMPLETO
✅ PASO 1: campo GemsHistory en PruningPointProof (commit)
✅ PASO 2: plomería - gemsHistoryStore conectado (struct+New+factory)
✅ PUNTO 1 (Build:285): buildPruningPointProof incluye el gemshistory
✅ PUNTO 2 (Validate:341): valida cordura (rechaza si excede topes MaxDiamante/Platino/Rodio, usa ErrGemsCapExceeded)
✅ PUNTO 3 (Apply:870): guarda el gemshistory verificado antes del CommitAllChanges
Todo compila (go build ./domain/consensus/...). Sin romper nada.

## PENDIENTE (próxima sesión, FRESCO)
1. SERIALIZACIÓN DE RED (protobuf) — que GemsHistory VIAJE entre nodos:
   - messages.proto / p2p.proto: agregar campo al mensaje PruningPointProof
   - Regenerar .pb.go con protoc (INSTALAR protoc primero, verificar)
   - Conversiones: domainconverters.go + p2p_pruning_point_proof.go
   - DELICADO: si sale mal rompe la red. Con calma + pruebas.
2. TESTS del flujo (build/validate/apply del gemshistory)
3. Probar en DEVNET antes de testnet/main.
4. NOTA: el "verificable total" (recalcular desde tx) NO es posible con
   solo headers. Enfoque actual = validación de cordura. El verificable
   fuerte requiere COMMITMENT EN HEADER (rediseño mayor, trabajo futuro).

## DIFICULTAD RESTANTE: serialización = MEDIA (mecánica pero delicada, toca protobuf de toda la red)

## ✅ SESIÓN COMPLETADA — CIRUGÍA CASI COMPLETA (rama pruning-verificable)

### FLUJO INTERNO (compila):
- PASO 1: campo GemsHistory en PruningPointProof
- PASO 2: plomería (gemsHistoryStore en el manager: struct+New+factory)
- PUNTO 1 (Build:285): el proof incluye el gemshistory del pruning point
- PUNTO 2 (Validate): valida cordura vía validateGemsHistorySanity()
- PUNTO 3 (Apply): guarda el gemshistory verificado

### SERIALIZACIÓN DE RED (compila todo el proyecto):
- p2p.proto: gemsHistory=2 + message GemsHistoryMessage (diamante/platino/rodio)
- protobuf regenerado con protoc v3.21.12 + protoc-gen-go v1.28.1
  (comando en generate.go; plugins en $GOPATH/bin, agregar al PATH)
- MsgGemsHistory en appmessage (p2p_msgpruningpointproof.go)
- 3 conversiones: protowire (p2p_pruning_point_proof.go) + domainconverters.go (x2)

### PRUEBAS (todas PASS):
- NIVEL 1 serialización (app/appmessage/p2p_gemshistory_serialization_test.go): 2 tests, el conteo viaja intacto
- NIVEL 2 validación (pruningproofmanager/gemshistory_sanity_test.go): 5 tests, acepta válidos/rechaza excesos/tope exacto ok

### FALTA (próxima sesión, FRESCO):
1. NIVEL 3: probar en DEVNET (integración real, 2 nodos, sincronización desde pruning point)
2. Merge a main (SOLO cuando devnet pase)
3. RECORDAR: el "verificable total" (recalcular desde tx) NO es posible con solo headers.
   Enfoque actual = validación de cordura. El verificable fuerte = COMMITMENT EN HEADER
   (rediseño mayor, trabajo futuro documentado).

### COMMITS en rama pruning-verificable: campo, plomería, Build, Validate, Apply,
### serialización, test nivel 1, test nivel 2. Todo compilando, 7 tests verdes.

## NOTA: test preexistente ExamplePayToAddrScript (txscript) FALLA
Verificado: falla IGUAL en main (sin la cirugia). NO es del pruning.
Relacionado con H-5 (campo Version sobrecargado por niveles de gema).
Es un pendiente conocido de la escalera, ajeno al pruning verificable.

## RANKING DE DIFICULTAD (corregido - Edu tiene razon)
LA MÁS DIFÍCIL DE TODO RUPIX = EL CAMBIO DE ALGORITMO (Autolykos), no el pruning.
Razones: toca el corazón del PoW (todos los bloques), sin ejemplo con GHOSTDAG,
irreversible en mainnet, la seguridad depende de que esté perfecto, requiere
criptografía de PoW + código C. Un error = red no produce bloques o hackeable.

El pruning fue el ENTRENAMIENTO: enseñó el method de tocar el motor (rama, tests,
devnet, radiografías). Ese method será el que salve el cambio de algoritmo.

Orden: 1)Algoritmo 2)Pruning total(commitment header) 3)Pruning cordura(hecho)
4)Endurecimiento 5)Hallazgos.
