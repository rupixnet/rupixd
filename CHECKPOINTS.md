# Checkpoints temporales en Rupix

🇬🇧 [English version](./CHECKPOINTS.en.md)

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

- Cada red tiene una lista `Checkpoints` (`domain/dagconfig/checkpoints.go`):
  pares `{BlueScore, Hash}`: el bloque canónico H y su blue score X.
- **Regla (v0.6.1, formulación para DAG):** al validar un encabezado (`checkCheckpoint`
  en `block_header_in_context.go`), todo bloque con blue score ≥ X + MergeDepth (3,600)
  debe tener a H en el pasado de alguno de sus padres. Si no, se rechaza con
  `ErrCheckpointMismatch`. Un hermano de H no se ve afectado (su blue score queda bajo
  el umbral); un bloque honesto posterior siempre tiene a H en su pasado; una historia
  alterna que no pasa por H se rechaza entera. Probado en `TestCheckpointDAG`, y se
  comprobó que la prueba falla si la regla se apaga.
- Un nodo que ya podó por debajo de H (no lo tiene) no aplica la regla: no puede.
- **Alcance hoy:** protege a nodos ya sincronizados. Un nodo que sincroniza desde cero
  recibe de un peer honesto a H como su propio punto de poda, así que la regla funciona
  por construcción; contra un peer hostil que sirve otro punto de poda con su propia
  prueba, la validación de la lista de puntos de poda contra los checkpoints llega en
  v0.6.2 (después de `ArePruningPointsInValidChain`). Por eso los checkpoints se publican
  **solo sobre puntos de poda**.
- `CheckpointsExpireDAAScore`: pasado ese DAA score, los checkpoints se ignoran.
  Es la fecha de caducidad, dentro del consenso, verificable.
- Lista vacía = sin efecto. Mainnet, simnet y devnet tienen la lista vacía; la testnet
  tiene el checkpoint #1 (abajo).

Probado en devnet (18 de septiembre de 2026): con el checkpoint correcto el nodo
acepta y mina encima; con un checkpoint falso, rechaza el bloque en ese DAA score
y la cadena no avanza.

## Cuándo se publica uno

Un checkpoint se publica solo sobre un bloque que ya tiene profundidad suficiente
(varios miles de bloques encima) y que los nodos externos ya tienen. Nunca sobre
bloques recientes. Cada checkpoint publicado se anuncia con: red, DAA score, hash,
fecha, y la versión del nodo que lo incluye.

**Historia (27-sep-2026), ya resuelta en v0.6.1:** la regla heredada era por DAA exacto; en un DAG habría partido la red (ver MEMORIA, "El checkpoint que habría partido la red"). Lo que decía entonces: un solo bloque en ese DAA. Rupix es un DAG: dos bloques hermanos pueden tener el mismo DAA score. El código rechaza *cualquier* bloque con el DAA del checkpoint y otro hash, así que un checkpoint puesto donde hay dos bloques invalidaría al hermano y a todo lo que lo incluye, y un nodo nuevo no podría sincronizar. Por eso, antes de publicar, se verifica con el nodo que en ese DAA hay exactamente un bloque. El arreglo de fondo (exigir el hash solo a bloques de la cadena) va en la próxima versión, con test.

## Cuándo se retiran

Los checkpoints se retiran cuando la red pueda sostenerse sola: hashrate externo
sostenido, varios nodos independientes, y semanas de estabilidad. Ese retiro se
hace poniendo `CheckpointsExpireDAAScore` en el consenso y anunciándolo. No hay
fecha fija: hay condiciones públicas.

## Registro de checkpoints publicados

| Red | DAA score | Hash | Fecha | Versión |
|---|---|---|---|---|
| testnet #4 | blue score 86,400 (DAA 86,399) | `7e2ece393c7d991c86e7ba915276cd85b5fc19e8647d5d197fa26bf116604fad` | 29-sep-2026 | v0.6.1 · caduca en DAA 2,000,000 |

*No confíes, verifica.*
