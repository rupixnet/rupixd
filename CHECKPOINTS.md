# Checkpoints temporales en Rupix

🇬🇧 [English version](./CHECKPOINTS.en.md)

## Qué son

Un checkpoint es un bloque canónico conocido, H, identificado por su hash y su blue
score X. La regla (v0.6.1, formulada para un DAG): todo bloque con blue score
≥ X + MergeDepth tiene que tener a H en su pasado. Cualquier nodo que reciba una
historia alternativa que no pase por H la rechaza, sin importar cuánto trabajo de
minado tenga. (Hasta v0.6.0 la regla heredada era "a cierto DAA score el bloque válido
es uno y solo uno"; en un DAG eso habría partido la red. Ver la historia más abajo.)

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
- **Alcance (v0.6.1):** protegía a nodos ya sincronizados. Un nodo que sincroniza desde cero
  recibe de un peer honesto a H como su propio punto de poda, así que la regla funciona
  por construcción; contra un peer hostil que sirve otro punto de poda con su propia
  prueba no había defensa.
- **Desde v0.6.2, también el nodo nuevo:** antes de importar el punto de poda que le sirve
  un peer (después de `ArePruningPointsInValidChain`), cada checkpoint activo por debajo
  de ese punto tiene que estar en la lista de puntos de poda recibida o ser un ancestro
  que el nodo conozca; si no, `ErrCheckpointMismatch` y el punto de poda no se importa
  (`TestNodoNuevoContraPeerHostil`). Por eso los checkpoints se publican **solo sobre
  puntos de poda**: así están en la lista de todo peer honesto.
- `CheckpointsExpireDAAScore`: pasado ese DAA score, los checkpoints se ignoran.
  Es la fecha de caducidad, dentro del consenso, verificable.
- Lista vacía = sin efecto. Mainnet, simnet y devnet tienen la lista vacía; la testnet
  tiene el checkpoint #1 (abajo).

Probado en devnet (18 de septiembre de 2026, con la regla anterior por DAA exacto):
con el checkpoint correcto el nodo acepta y mina encima; con uno falso, rechaza y la
cadena no avanza. La regla actual por blue score la prueba `TestCheckpointDAG` (v0.6.1):
un hermano tardío de H entra, una historia sin H se rechaza, y la prueba falla si la
regla se apaga.

## Cuándo se publica uno

Un checkpoint se publica solo sobre un bloque que ya tiene profundidad suficiente
(varios miles de bloques encima) y que los nodos externos ya tienen. Nunca sobre
bloques recientes. Cada checkpoint publicado se anuncia con: red, DAA score, hash,
fecha, y la versión del nodo que lo incluye. (El registro de abajo da el blue score, que es lo que usa la regla; el DAA score se anota como referencia.)

**Historia (27-sep-2026), ya resuelta en v0.6.1:** la regla heredada era por DAA exacto; en un DAG habría partido la red (ver MEMORIA, "El checkpoint que habría partido la red"). Lo que decía entonces: un solo bloque en ese DAA. Rupix es un DAG: dos bloques hermanos pueden tener el mismo DAA score. El código rechaza *cualquier* bloque con el DAA del checkpoint y otro hash, así que un checkpoint puesto donde hay dos bloques invalidaría al hermano y a todo lo que lo incluye, y un nodo nuevo no podría sincronizar. Por eso, antes de publicar, se verifica con el nodo que en ese DAA hay exactamente un bloque. El arreglo de fondo (H en el pasado de todo bloque con blue score ≥ X + MergeDepth) salió en v0.6.1, con `TestCheckpointDAG`.

## Qué pasa cuando la poda avanza más allá del checkpoint

Verificado el 30-sep-2026: el punto de poda de la testnet pasó de H (blue score 86,400) a blue score 172,800, y el seed siguió teniendo a H como bloque de cadena y aplicando la regla. Los puntos de poda pasados se conservan en el nodo aunque la poda avance; por eso los checkpoints se publican solo sobre puntos de poda: nunca desaparecen del nodo que los valida.

## Renovación: nunca hay un hueco entre checkpoints

Cada checkpoint tiene caducidad (`CheckpointsExpireDAAScore`). La regla desde el 30-sep-2026: **el siguiente checkpoint se publica antes de que caduque el anterior**, siempre sobre un punto de poda más reciente y con una caducidad nueva, en una versión del nodo anunciada con tiempo. Un hueco entre checkpoints sería una ventana de ataque anunciada de antemano; por eso no se permite. Si por alguna razón no hay checkpoint nuevo listo antes de la caducidad, se publica una versión que solo extiende la caducidad del vigente. La red solo se queda sin candado cuando se retire a propósito (sección siguiente), nunca por descuido. Para el #1 (caduca en DAA 2,000,000): el #2 se publica a más tardar en el DAA 1,700,000 (~3.5 días de testnet antes).

**Cómo se prepara el siguiente (desde el 6-oct-2026):** `tools/checkpoint-propuesto.sh [pruningPointHash de un nodo externo]` lee el punto de poda actual del nodo, comprueba la profundidad (≥ 100,000 bloques encima) y si el nodo externo tiene el mismo punto de poda, e imprime las líneas exactas para `checkpoints.go`, `checkpoints_test.go` y esta tabla. No cambia nada por sí solo. La caducidad es **una sola para la lista** (`CheckpointsExpireDAAScore`): al publicar un checkpoint nuevo se mueve para toda la lista, así que los anteriores siguen vigentes hasta la nueva caducidad. No quita nada: cada checkpoint anterior está en el pasado del siguiente, y un nodo que sincroniza desde cero los comprueba todos (v0.6.2). Si al momento de publicar no hay ningún nodo externo encendido con el que cotejar, el checkpoint se publica igual (un hueco es peor) y la tabla dice que nadie externo lo confirmó.

## Cuándo se retiran

Los checkpoints se retiran cuando la red pueda sostenerse sola: hashrate externo
sostenido, varios nodos independientes, y semanas de estabilidad. Ese retiro se
hace poniendo `CheckpointsExpireDAAScore` en el consenso y anunciándolo. No hay
fecha fija: hay condiciones públicas.

## Registro de checkpoints publicados

| Red | Blue score (y DAA) | Hash | Fecha | Versión |
|---|---|---|---|---|
| testnet #4 | blue score 86,400 (DAA 86,399) | `7e2ece393c7d991c86e7ba915276cd85b5fc19e8647d5d197fa26bf116604fad` | 29-sep-2026 | v0.6.1 · caduca en DAA 2,000,000 |

*No confíes, verifica.*
