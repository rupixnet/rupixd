# 💎 Cómo forjar tu primera gema en Rupix

🇬🇧 [English version](./FORGE-GUIDE.md)

Guía para crear tu primer Diamante quemando Gold. Esto es la
esencia de Rupix: destruir para crear algo más escaso.

> **Requisito:** necesitas tener **Gold** (RUPIX) en tu wallet.
> Consíguelo minando (ver la guía de minado) o pídele a alguien
> que te mande un poco. Para forjar 1 Diamante necesitas **al
> menos 10 Gold** (10 se queman) + un poquito extra para la
> comisión de la transacción.

> **⚠️ Importante — guarda Gold aparte:** mover o transferir una gema
> también cuesta una pequeña comisión en Gold. Si forjas una gema y te
> quedas SIN nada de Gold, tu gema queda quieta: no la podrás mover ni
> transferir hasta que tengas un poco de Gold otra vez. **No la perdiste**
> —sigue siendo tuya y visible en la cadena—, solo necesita algo de Gold
> para moverse. Consejo: guarda siempre un poco de Gold aparte.

> **⚠️ Windows vs Linux/Mac:** los comandos de abajo empiezan con `./` (Linux/Mac).
> En **Windows PowerShell**, usa `.\` en vez de `./` — por ejemplo: `.\rupixd.exe`.
> Sin el `./` o `.\`, el sistema no encuentra el programa. (¡Gracias a JC por el aporte!)
---

## Paso 1 — Ten tu wallet y daemon corriendo

Si aún no tienes wallet, créala (una sola vez):
```
./rupixwallet --testnet create
```
Arranca el daemon del wallet (déjalo corriendo en una ventana):
```
./rupixwallet --testnet start-daemon
```

---

## Paso 2 — Revisa tu Gold

En otra ventana, mira cuánto Gold tienes:
```
./rupixwallet --testnet balance
```
Necesitas al menos 11 RUPIX de Gold para forjar un Diamante (10 que se queman + margen para la comisión; la wallet lo comprueba antes de mandar). Para Platino, Rodio y Kings: las 10 gemas del nivel anterior y al menos 2 RUPIX de Gold para la comisión
(10 se queman + un poco para la comisión).

---

## Paso 3 — Crea una dirección para tu gema

La gema "nace" en una dirección tuya. Crea una (o usa una que ya tengas):
```
./rupixwallet --testnet new-address
```
Copia la dirección que empieza con `rupixtest:...`

---

## Paso 4 — ¡Forja tu Diamante!

> **Antes de forjar:** cada nivel se abre en su halving. En la testnet: Diamante en el DAA 100,000,
> Platino en 200,000, Rodio en 300,000 y Kings en 400,000. Antes de ese punto la red rechaza la forja.
> Revisa en qué DAA va la red con `./rupixctl --testnet GetBlockDagInfo` (campo `virtualDaaScore`)
> y deja unos bloques de margen.
>
> `--level`: 1 = Diamante, 2 = Platino, 3 = Rodio, 4 = Kings. El Gold no se forja: se mina.
>
> Desde v0.6.1 la wallet te pide la contraseña en pantalla: no la escribas en la línea de
> comandos (quedaría en el historial). Si usas v0.6.0 todavía, actualiza antes de forjar.

```
./rupixwallet --testnet forge --level=1
Vas a forjar 1 Diamante en rupixtest:qq...tu_direccion
Se queman 10 Gold PARA SIEMPRE, mas una comision pequena.
Confirmas? Escribe 'si' para continuar: si
Contrasena: (la escribes aquí, no se ve)
Ascenso forjado: gema Diamante creada.
```

- `--level=1` → Diamante (el primer nivel de gema)
- `--gem-address=` es **opcional**: si no lo pones, la gema nace en tu primera dirección (la del paso 3)
- Antes de pedir la contraseña, la wallet consulta tu nodo: si el nivel no está abierto o falta Gold, te lo dice en palabras y no manda nada
- La contraseña se pide al momento, siempre; desde v0.6.3 ya no existe la bandera `--password` en ningún comando (quedaba en el historial)

Esto **quema 10 Gold para siempre** y crea **1 Diamante**. La quema
queda grabada en la blockchain, visible para todos, irreversible.

Los niveles (para más adelante):
- `--level=1` → Diamante (quema 10 Gold)
- `--level=2` → Platino (quema 10 Diamantes)
- `--level=3` → Rodio (quema 10 Platinos)
- `--level=4` → Kings (quema 10 Rodios)

---

## Paso 5 — Mira tu gema

```
./rupixwallet --testnet gems
```
Verás tu inventario. ¡Ahí está tu Diamante! Eres de los primeros
en forjar una gema en Rupix. 💎

---

## ¿Qué acabas de hacer?

Creaste una **prueba criptográfica de que destruiste valor real**.
Tu Diamante demuestra que quemaste 10 Gold, para siempre. No es
una imagen ni un número inventado: es escasez verificable, grabada
en la cadena. Solo 2,100,000 Diamantes existirán jamás.

Cada nivel superior es 10 veces más raro. Un King (nivel 4) exige
quemar 10,000 Gold en total. Solo 2,100 Kings existirán en toda
la historia de Rupix.

---

*Rupix — el dinero que solo se consume. No confíes. Verifica — desde el génesis.*
🔗 rupix.network · github.com/rupixnet/rupixd · explorer.rupix.network

## Qué cambió en la wallet (v0.6.2) y por qué

Estos cambios existen porque vimos a personas reales tropezar con la wallet. Cada uno evita un error concreto.

| Antes | Ahora | Por qué |
|---|---|---|
| `forge` mandaba la transacción y esperaba a que la red la rechazara | **Revisa antes:** cuenta tus gemas y tu Gold, le pregunta al nodo en qué bloque va y si el nivel está abierto. Si falta algo, lo dice en palabras y **no manda nada** | La red rechazó un Platino a JC el 28-sep porque el nivel no estaba abierto; con esto lo habría sabido antes de intentar |
| Forjar era inmediato e irreversible | **Resume y pregunta:** "Vas a forjar 1 Platino… se queman 10 Diamantes PARA SIEMPRE… ¿Confirmas?" Continúa con `si`, `s` o `yes`; cualquier otra cosa cancela | Lo que se quema no vuelve. Un dedo de más no debe costarte 10 gemas |
| `send` firmaba y mandaba sin más | Dice cuánto, a quién y **en cuántas transacciones** (un envío grande se parte en lotes), y pregunta | 10,000 RUPIX salen en 232 transacciones; hay que saberlo antes de firmar |
| `--gem-address` obligatorio | Opcional: si no lo das, la gema nace en tu primera dirección | Un paso menos para quien empieza |
| `--password=` en la línea de comandos | Se pide en pantalla, sin eco | Quedaba en el historial y en las capturas de pantalla |
| Errores del nodo en crudo (`nivel 2 bloqueado…`) | Explicados en palabras, con el original abajo entre paréntesis | Para saber qué hacer, no solo qué pasó |
| Solo español | **Español o inglés**, cambiable cuando quieras | Rupix es para todos |

**Elegir idioma:** `./rupixwallet language en` lo deja guardado (o `es`). Para una sola vez: `--lang=en` antes del comando. También vale la variable `RUPIX_LANG`. Si no dices nada, usa el idioma del sistema (`LANG`) y, si no hay, español.

**Para scripts** que no pueden contestar preguntas: `--yes` en `forge`, `send` y `transfer-gem`.

Nada de esto toca el consenso: la red sigue revisando todo igual. La wallet solo te avisa antes.
