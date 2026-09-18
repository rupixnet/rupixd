# Checkpoints temporales en Rupix

## Qué son

Un checkpoint es un bloque canónico conocido: a cierta altura de la red (DAA score),
el bloque válido es uno y solo uno, identificado por su hash. Cualquier nodo que
reciba una cadena alternativa que no pase por ese bloque la rechaza, sin importar
cuánto trabajo de minado tenga.

## Por qué Rupix los usa

Una red joven tiene poco hashrate. Con poco hashrate, un atacante con hardware
suficiente podría reescribir la historia (ataque del 51%): deshacer transacciones,
gastar dos veces, borrar forjas. Los checkpoints hacen que eso sea imposible para
todo lo anterior al último checkpoint: el atacante solo podría afectar el tramo
posterior, y ese tramo es corto si los checkpoints se publican con regularidad.

Es una defensa **temporal** y es una forma de **centralización declarada**: quien
publica el checkpoint decide qué historia es la canónica. Rupix lo hace a la vista
de todos, con la regla escrita, y con compromiso de retirarlo.

## Cómo funciona (en el código)

- Cada red tiene una lista `Checkpoints` en `domain/dagconfig/params.go`:
  pares `{DAAScore, Hash}`.
- Al validar un encabezado (`checkCheckpoint` en `block_header_in_context.go`),
  si el DAA score del bloque coincide con un checkpoint y su hash no es el
  canónico, el bloque se rechaza con `ErrCheckpointMismatch`.
- Un bloque en un DAA score sin checkpoint pasa sin más.
- `CheckpointsExpireDAAScore`: pasado ese DAA score, los checkpoints se ignoran.
  Es la fecha de caducidad, dentro del consenso, verificable.
- Lista vacía = sin efecto. Hoy todas las redes tienen la lista vacía.

Probado en devnet (18 de septiembre de 2026): con el checkpoint correcto el nodo
acepta y mina encima; con un checkpoint falso, rechaza el bloque en ese DAA score
y la cadena no avanza.

## Cuándo se publica uno

Un checkpoint se publica solo sobre un bloque que ya tiene profundidad suficiente
(varios miles de bloques encima) y que los nodos externos ya tienen. Nunca sobre
bloques recientes. Cada checkpoint publicado se anuncia con: red, DAA score, hash,
fecha, y la versión del nodo que lo incluye.

## Cuándo se retiran

Los checkpoints se retiran cuando la red pueda sostenerse sola: hashrate externo
sostenido, varios nodos independientes, y semanas de estabilidad. Ese retiro se
hace poniendo `CheckpointsExpireDAAScore` en el consenso y anunciándolo. No hay
fecha fija: hay condiciones públicas.

## Registro de checkpoints publicados

| Red | DAA score | Hash | Fecha | Versión |
|---|---|---|---|---|
| — | — | — | — | (ninguno aún) |

*No confíes, verifica.*
