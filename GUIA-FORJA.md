# 💎 Cómo forjar tu primera gema en Rupix

Guía para crear tu primer Diamante quemando Gold. Esto es la
esencia de Rupix: destruir para crear algo más escaso.

> **Requisito:** necesitas tener **Gold** (RUPIX) en tu wallet.
> Consíguelo minando (ver la guía de minado) o pídele a alguien
> que te mande un poco. Para forjar 1 Diamante necesitas **al
> menos 10 Gold** (10 se queman) + un poquito extra para la
> comisión de la transacción.

> **⚠️ Windows vs Linux/Mac:** los comandos de abajo empiezan con `./` (Linux/Mac).
> En **Windows PowerShell**, usa `.\` en vez de `./` — por ejemplo: `.\rupixd.exe`.
> Sin el `./` o `.\`, el sistema no encuentra el programa. (¡Gracias a Coco por el aviso!)
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
Necesitas al menos ~10.5 RUPIX para forjar un Diamante
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

```
./rupixwallet --testnet forge --level=1 --gem-address=TU_DIRECCION --password=TU_CLAVE
```

- `--level=1` → Diamante (el primer nivel de gema)
- `--gem-address=` → la dirección donde nace la gema (la del paso 3)
- `--password=` → la contraseña de tu wallet

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
