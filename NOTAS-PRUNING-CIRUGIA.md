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
