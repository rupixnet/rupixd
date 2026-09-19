# 🔷 Cómo unirte a la testnet de Rupix

Guía sencilla para conectar tu nodo a la red de prueba de Rupix.
No necesitas ser experto — sigue los pasos.

> **Rupix testnet** es una red de PRUEBA. Las monedas (RUPIX) de testnet
> NO tienen valor real. Es para probar, aprender y ayudar a fortalecer
> la red antes de mainnet. ¡Gracias por sumarte!

> **⚠️ Windows vs Linux/Mac:** los comandos de abajo empiezan con `./` (Linux/Mac).
> En **Windows PowerShell**, usa `.\` en vez de `./` — por ejemplo: `.\rupixd.exe`.
> Sin el `./` o `.\`, el sistema no encuentra el programa. (¡Gracias a JC por el aporte!)
---

## Paso 1 — Descarga Rupix

Ve a las descargas oficiales:
**https://github.com/rupixnet/rupixd/releases**

Descarga el archivo de tu sistema:
- **Windows:** `rupix-v0.5.1-win64.zip`
- **Mac:** `rupix-v0.5.1-osx.zip`
- **Linux:** `rupix-v0.5.1-linux.zip`

Descomprime el archivo. Dentro encontrarás 4 programas:
`rupixd` (el nodo), `rupixctl` (control), `rupixwallet` (billetera), `rupixminer` (minero).

---

## Paso 2 — Arranca tu nodo y conéctate a la red

Abre una terminal en la carpeta donde descomprimiste Rupix, y corre:

**Linux / Mac:**
```
./rupixd --testnet --utxoindex --addpeer=178.104.69.148:17211
```

**Windows:**
```
.\rupixd.exe --testnet --utxoindex --addpeer=178.104.69.148:17211
```

Tu nodo se conectará al nodo semilla de Rupix y empezará a
**sincronizar** (descargar la cadena). Verás muchos mensajes —
es normal. Cuando veas que va alcanzando el bloque más reciente,
¡ya estás en la red! 🎉

---

## Paso 3 — Verifica que estás conectado

En OTRA terminal (deja el nodo corriendo en la primera), corre:

**Linux / Mac:**
```
./rupixctl --testnet GetBlockDagInfo
```

**Windows:**
```
.\rupixctl.exe --testnet GetBlockDagInfo
```

Deberías ver:
- `networkName: "rupix-testnet"` (estás en la red correcta)
- `blockCount` subiendo (tu nodo está sincronizando)

Compara tu `blockCount` con el del explorador oficial
(**https://explorer.rupix.network**). Cuando se acerquen, estás al día.

---

## Paso 4 (opcional) — Mina Rupix

> **Sobre el minado:** Rupix usa **RupixHeavyHash**, un algoritmo propio
> derivado del motor de Kaspa (kHeavyHash; Rupix es un fork de kaspad) con nuestra propia semilla.
> Puedes minar desde tu CPU o GPU, y en testnet tu computadora puede ganar
> bloques. Lo importante ahora es correr tu **nodo** (pasos 2-3); minar es un
> extra que también suma a la red.
>
> RupixHeavyHash impide que las máquinas industriales fabricadas para Kaspa
> minen Rupix — un arranque justo para que todos empiecen parejos. Honesto:
> es una ventaja de meses, no independencia permanente; un FPGA puede
> reprogramarse. Es un arranque justo, no una barrera eterna.
>
> **Nota sobre el archivo de llaves:** si arrancas el daemon con
> `--keys-file=RUTA`, el comando `send` también debe llevar el mismo
> `--keys-file=RUTA` (firma localmente con esas llaves). Los demás comandos
> (`new-address`, `balance`, `gems`, `forge`) NO lo llevan: consultan al daemon.
> Si no usas `--keys-file` en ningún lado, todo usa el archivo por defecto.

¿Quieres ayudar a minar la testnet? Primero crea una dirección:

```
./rupixwallet --testnet create
```
(Te pedirá una contraseña. Guárdala. Luego crea una dirección con:)
```
./rupixwallet --testnet start-daemon
./rupixwallet --testnet new-address
```

Copia tu dirección (empieza con `rupixtest:...`) y arranca el minero:

```
./rupixminer --testnet --miningaddr=TU_DIRECCION_AQUI
```

¡Estarás minando Gold de testnet y ayudando a fortalecer la red!

---

## ¿Problemas?

- **No sincroniza / no conecta:** revisa tu internet y que escribiste
  bien la dirección del peer (`178.104.69.148:17211`).
- **"address already in use":** ya tienes un nodo corriendo; ciérralo antes.
- Escríbeme y lo resolvemos juntos.

---

*Rupix testnet — el laboratorio público. Todos podemos minar, todos podemos verificar.*
**No confíes. Verifica — desde el génesis.**
🔗 rupix.network · github.com/rupixnet/rupixd · explorer.rupix.network
