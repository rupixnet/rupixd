# Operación del nodo semilla / Seed node operations

🇲🇽 El seed de la testnet corre como tres servicios systemd con reinicio automático. Si el nodo muere (OOM, crash, reinicio del servidor), se levanta solo en segundos. Lección del 21-sep-2026: el nodo murió por OOM a las 04:31 y nadie lo supo hasta las 19:00 — 14 horas sin seed. Nunca más.

🇬🇧 The testnet seed runs as three systemd services with automatic restart. If the node dies (OOM, crash, server reboot), it comes back on its own within seconds. Lesson from 21-Sep-2026: the node was OOM-killed at 04:31 and nobody knew until 19:00 — 14 hours without a seed. Never again.

## Servicios / Services

| Servicio | Qué es | Depende de |
|---|---|---|
| `rupixd-testnet` | el nodo (RPC 17210, P2P 17211) | red |
| `rupixwallet-testnet` | daemon de la wallet del servidor (8082) | rupixd |
| `rupixminer-testnet` | minero | rupixd |

## Comandos / Commands

- Estado / status: `systemctl status rupixd-testnet rupixwallet-testnet rupixminer-testnet`
- Reiniciar nodo (daemon y minero lo siguen): `systemctl restart rupixd-testnet`
- Log del servicio: `journalctl -u rupixd-testnet -n 50`
- Log del nodo: `tail -f /root/rupix-testnet.log`
- Bloque actual: `rupixctl -s 127.0.0.1:17210 GetBlockDagInfo`

## Actualizar el binario / Update the binary

1. `go build -o /usr/local/bin/rupixd .` (y wallet, miner, ctl)
2. `systemctl restart rupixd-testnet` — la cadena en disco no se toca; daemon y minero se reinician solos.

## Memoria / Memory

El servidor tiene 7.7 GB. rupixd usa ~2.7 GB. **No correr `go test ./...` completo en este host** mientras el seed corre: fue la causa del OOM del 21-sep. Tests pesados en otra máquina o con `go test -p 1`.

## Reglas / Rules

- Si un proceso está `activating` en loop: `journalctl -u <servicio> -n 30` dice por qué.
- Nunca `kill -9` a mano: `systemctl stop/restart`.
- Tras un reinicio del servidor, los tres arrancan solos (`enabled`).

## Alarma de bloques descalificados / Disqualified-block alarm (auditor, 23-sep)

> **Esta es la v1 y quedó reemplazada el 26-sep:** caminaba por la cadena seleccionada y no podía ver un bloque descalificado. Ver más abajo "Corrección de la alarma" y "Alarma v2.1". / *This is v1, replaced on 26-Sep: it could not see disqualified blocks. See the correction sections below.*

🇲🇽 `/root/rupix-monitor-descalificados.sh` revisa cada 10 min (cron) los últimos 200 bloques de la cadena y cuenta los descalificados (isChainBlock:false). Un bloque descalificado en una red donde el fundador tiene casi todo el hashrate NO es ruido: es un bug de la costura minero/validador hasta que se demuestre lo contrario (así estuvo escondido 10 días el bug del King). Revisar: `tail /root/rupix-descalificados.log`. Si dice ALARMA, investigar antes que nada.

🇬🇧 `/root/rupix-monitor-descalificados.sh` runs every 10 min (cron), checks the last 200 chain blocks, counts disqualified ones (isChainBlock:false). A disqualified block on a network where the founder holds nearly all hashrate is NOT noise — it's a miner/validator seam bug until proven otherwise. Check: `tail /root/rupix-descalificados.log`.

## Devnets de prueba (regla desde 26-sep-2026)

- En el seed solo corre la testnet. Las devnets se prenden cuando se van a usar y se apagan al terminar.
- Por qué: el 26-sep se encontró la devnet de las pruebas de v0.6.0 prendida desde hacía 2 días. Usaba 1.8 GB de RAM (el seed ya tuvo un OOM) y tenía su P2P (17651) abierto a internet. Al apagarla, la RAM disponible subió de 3.2 a 5.0 GB.
- Datos en disco (apagar no los borra): `/root/rupix-devnet-v06` (pruebas de v0.6.0), `/root/rupix-devnet-cp` (checkpoints), `/root/rupix-devnet-algo` y `/root/wallet-devnet-algo` (pruebas del algoritmo). Sus logs `.log` están en `/root`.
- Para usar una devnet: compilar el binario fuera de `/tmp` (se borra al reiniciar) y arrancar con P2P y RPC solo locales:
  `rupixd --devnet --appdir=<carpeta> --listen=127.0.0.1:17651 --rpclisten=127.0.0.1:17350`
- Al terminar: `kill <pid>` y comprobar con `ss -tlnp | grep rupixd` que solo quede la testnet (17211 y 17210).

## Corrección de la alarma de descalificados (26-sep-2026)

- **La v1 no servía.** Caminaba hacia atrás por la cadena seleccionada desde la punta. Todo bloque en ese camino está en la cadena por definición, y un bloque descalificado es justo el que queda fuera. Nunca podía encontrarlo: dijo "OK" 134 veces sin poder sonar. También era falso que "habría cazado el bug del King".
- **La v2:** el nodo registra cada descalificación en su log (`resolve_block_status.go`, nivel Warn: `BLOQUE DESCALIFICADO <hash>: <motivo>`). La alarma lee solo lo que el log agregó desde la revisión anterior (posición guardada en `/root/.rupix-alarma-offset`) y escribe en `/root/rupix-descalificados.log`. Si el log no creció, escribe ERROR: nunca un OK sin datos.
- El nodo escribe en `/root/rupix-testnet/<red>/logs/rupixd.log`, no en el journal.
- **Probada** con un log falso: suena. **Pendiente:** provocar un bloque descalificado en devnet y verla sonar de punta a punta.
- La v1 queda en `/root/rupix-monitor-descalificados.v1-ciego.sh`, como registro.

## Dónde vive cada cosa (26-sep-2026)

- **rupix.network** se sirve desde **GitHub Pages**, con el repo `rupixnet/rupix-website`. Se publica solo al hacer push y tarda 1–2 minutos. Un 404 justo después del push es normal.
- **explorer.rupix.network** apunta al seed (178.104.69.148). nginx escucha en 80/443. La API la sirve `rupixexplorer` en `127.0.0.1:8090`, conectado al nodo por RPC (`127.0.0.1:17210`). La página del explorador está en `/root/rupix-explorer-web/` (falta documentar cómo la sirve nginx).
- **Nodo:** `rupixd` en 17211 (P2P, abierto) y 17210 (RPC, solo local). Wallet daemon en `127.0.0.1:8082`.

## Alarma v2.1: rotación del log (27-sep-2026)

- El nodo rota su log cada 100 MB y guarda 8 archivos, sin comprimir (`infrastructure/logger/backend.go`). Al rotar, `rupixd.log` pasa a otro nombre y conserva su inode.
- La alarma guarda **inode + posición** en `/root/.rupix-alarma-offset`. Si el inode cambió, busca el archivo viejo por su inode, termina de leerlo desde donde iba y luego lee el nuevo desde el principio. Si no encuentra el viejo, escribe ERROR.
- Probada con rotaciones simuladas: detecta un descalificado que quedó en el archivo viejo justo antes de rotar; si falta el archivo viejo, dice ERROR; después se recupera sola.
- Lo pidió el auditor al revisar `195940d`.

## La página del explorer (27-sep-2026)

- `rupixexplorer` sirve la carpeta `/root/rupix-explorer-web` (su opción `--webdir`, que por defecto es esa).
- La fuente de la página es `cmd/rupixexplorer/web/index.html`, en el repo. Se cambia ahí, se commitea y luego se copia: `cp /root/rupixd/cmd/rupixexplorer/web/index.html /root/rupix-explorer-web/index.html`. No hace falta reiniciar nada.
- El 27-sep la copia del repo estaba vieja (halving de testnet en 10,000; el real es 100,000) y la viva era la correcta. Se sincronizó. Antes de copiar, revisa con `diff` que las dos coincidan.

## Todo lo del seed, versionado (27-sep-2026)

- Los servicios de systemd, la alarma, el cron y la configuración de nginx del explorer están copiados en la carpeta `ops/` del repo (ver `ops/README.md`). Sin secretos. Si el seed se pierde, se reconstruye desde ahí.

- 27-sep: el servicio `rupixd.service` (v0.3.0, misma carpeta y puerto que el nodo actual) seguía habilitado; al reiniciar habría chocado con el nodo. Deshabilitado, no borrado: está en `/root/rupixd.service.v030-retirado` y en `ops/retirados/`.

- 27-sep: el explorer corría lanzado a mano desde el 9-sep, fuera de systemd, y no habría vuelto tras un reinicio. Ahora lo lleva `rupix-explorer.service` (Restart=always, después de `rupixd-testnet`).

- 27-sep: los logs de salida no se rotaban (256 MB el del minero). Ahora se rotan comprimidos y se guardan todos (`/etc/logrotate.d/rupix`, copia en `ops/logrotate/`).

## Dirección de minado del seed (28-sep-2026)

- El minero cobró en `rupixtest:qq740lal…jd27l` del DAA 2 al 92,569 y en `rupixtest:qp4y8vnk…cpylzw` desde el 92,581 (testnet #4). Las dos son de la misma wallet del seed. Verificable con `GetUtxosByAddresses` (primer y último `blockDaaScore` de cada dirección).
- Envíos de más de ~820 RUPIX desde la wallet del minero se cortan a los 2 minutos (ver CONTEXTO, 28-sep). Mandar en partes hasta el arreglo.

## Consultar gemas de una dirección desde el nodo (29-sep-2026)

`rupixctl --testnet GetUtxosByAddresses <direccion>` (la dirección va como parámetro suelto; varias, separadas por coma). Cada entrada trae `outpoint.transactionId`, `utxoEntry.scriptPublicKey.version` (0 Gold, 1 Diamante, 2 Platino, 3 Rodio, 4 Kings) y `utxoEntry.blockDaaScore` (bloque donde nació). Sirve para verificar una forja de cualquiera sin tocar su wallet, y para comparar dos nodos entre sí (misma lista = mismo estado).

## Actualizar el seed a una versión nueva (hecho por primera vez con v0.6.1, 29-sep-2026)

1. En `/root/rupixd`, `main` al día y compilado: `go build -o /root/bin-vXYZ/ . ./cmd/rupixctl ./cmd/rupixwallet ./cmd/rupixminer`.
2. Copiar los binarios viejos a `/root/bin-anterior/` (nada se borra) y los nuevos a `/usr/local/bin/`.
3. `systemctl restart rupixd-testnet rupixminer-testnet rupixwallet-testnet rupix-explorer`; esperar 30 s.
4. Verificar: `rupixd --version`, `rupixctl --testnet GetInfo` (versión del servidor), `GetBlockDagInfo` (DAA avanza, `isSynced`), `journalctl -u rupixd-testnet -n 20`.
5. La release en GitHub (tag + CI) es lo que actualiza a la comunidad; el seed puede ir un paso adelante, nunca atrás.

## Verificar una release (30-sep-2026)

Después de publicar una release, en el seed: `sh /root/rupixd/tools/verificar-binarios.sh vX.Y.Z`. Debe decir `OK` con cuatro `IGUAL`. Si difiere, no se anuncia la release hasta entender por qué (versión de Go distinta a la del CI, flags cambiados, o algo peor). El CI compila con Go 1.26.6 y CGO activo con enlace estático; el seed debe tener la misma versión de Go.

## Levantar un seed desde cero / Setting up a seed from scratch (7-oct-2026)

🇲🇽 Esto es lo que hace falta para que **cualquier persona** levante un nodo semilla de la testnet igual al nuestro, de arriba a abajo. Es lo que seguirá el segundo seed. Un seed es solo un nodo con el puerto P2P abierto a internet y una dirección fija: el minero y la wallet son opcionales.

🇬🇧 This is what **anyone** needs to run a testnet seed node identical to ours, top to bottom. It is what the second seed will follow. A seed is just a node with its P2P port open to the internet and a fixed address: miner and wallet are optional.

### 1. Máquina / Machine
- 4 núcleos, 8 GB de RAM, 80 GB de disco SSD, IP pública fija. Ubuntu 24.04 LTS. (El seed 1: 4 núcleos, 7.6 GB, 75 GB; la cadena ocupa ~7.5 GB al DAA 1,000,000 y crece ~0.3 GB/día con poda.)
- Proveedor y país **distintos** del seed 1, a propósito: la red no debe depender de una empresa ni de una jurisdicción.
- 4 cores, 8 GB RAM, 80 GB SSD, fixed public IP, Ubuntu 24.04 LTS. Different provider and country from seed 1, on purpose.

### 2. Firewall
```
ufw allow 22/tcp        # SSH
ufw allow 17211/tcp     # P2P de Rupix testnet, abierto / open
ufw deny  17210         # RPC: nunca a internet / never to the internet
ufw enable
```
(Mainnet, cuando exista: P2P 17111, RPC 17110.)

### 3. Binarios verificados / Verified binaries
```
cd /root
V=v0.6.4
curl -sSLO https://github.com/rupixnet/rupixd/releases/download/$V/rupix-$V-linux.zip
curl -sSLO https://github.com/rupixnet/rupixd/releases/download/$V/rupix-$V-linux.zip.sha256
sha256sum -c rupix-$V-linux.zip.sha256          # debe decir OK; si no, no sigas / must say OK
mkdir -p /root/$V && unzip -q rupix-$V-linux.zip -d /root/$V
cp $(find /root/$V -type f \( -name rupixd -o -name rupixctl -o -name rupixwallet -o -name rupixminer \)) /usr/local/bin/
rupixd --version
```
Quien quiera ir más lejos recompila desde el tag y compara (`tools/verificar-binarios.sh $V`, Go 1.26.6).

### 4. El servicio del nodo / The node service
`/etc/systemd/system/rupixd-testnet.service` (es el del seed 1, con dos cambios: la IP, y sin `--allow-submit-block-when-not-synced`, que solo tenía sentido cuando no había nadie a quien sincronizarse):
```
[Unit]
Description=Rupix testnet seed node
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/rupixd --testnet --appdir=/root/rupix-testnet --utxoindex --listen=0.0.0.0:17211 --externalip=<IP PUBLICA DE ESTE SEED> --rpclisten=127.0.0.1:17210 --addpeer=178.104.69.148:17211
Restart=always
RestartSec=5
StandardOutput=append:/root/rupix-testnet.log
StandardError=append:/root/rupix-testnet.log
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
```
```
systemctl daemon-reload && systemctl enable --now rupixd-testnet
sleep 30 && rupixctl --testnet GetBlockDagInfo | grep -E "virtualDaaScore|pruningPointHash"
rupixctl --testnet GetConnectedPeerInfo | grep -c '"address"'     # >= 1
```
Sincroniza desde cero contra el seed 1 (unas horas). Cuando `virtualDaaScore` alcance al del seed 1 y `pruningPointHash` coincida, es un seed. Esa coincidencia se anota en `CHECKPOINTS.md` como confirmación del checkpoint vigente.

### 5. Opcional: minero y wallet / Optional: miner and wallet
Los `.service` del seed 1 (`rupixminer-testnet`, `rupixwallet-testnet`) están en este archivo arriba en espíritu; el minero lleva `--miningaddr=<tu direccion>` y, en un seed 2, **sin** `--mine-when-not-synced`. La wallet del servicio corre con su propio `keys.json`; la contraseña se pide al arrancar o va en un archivo solo legible por root. Un seed no necesita ninguno de los dos.

### 6. Alarma / Alarm
`/root/rupix-monitor-descalificados.sh` cada 10 min por cron (sección "Alarma v2.1"). Copiarlo del seed 1 y registrar la línea `*/10 * * * * /root/rupix-monitor-descalificados.sh` con `crontab -e`.

### 7. Anunciarlo / Announcing it
- `GUIA-TESTNET.md` / `TESTNET-GUIDE.md`, paso 2: segundo `--addpeer`.
- `dagconfig/params.go` tiene `DNSSeeds: seed.rupix.network`: ese nombre debe resolver a **todas** las IPs de seeds (un registro A por seed). Si no resuelve, los nodos sin `--addpeer` no encuentran la red. **Pendiente de verificar el 7-oct: si `seed.rupix.network` existe en el DNS.**
- `SUCESION.md`: tachar "segundo seed" con fecha. `MAINNET.md`: criterio N1.
- Quién lo opera y cómo se le avisa de una release: en `ACCESOS-RUPIX.md` (privado), nunca aquí.

### 8. Mantenimiento que ya nos mordió / Maintenance that already bit us
- **Disco:** el 7-oct el seed 1 estaba al 82 % con la cadena en 7.5 GB; el resto era caché de Go y archivos viejos. Revisar cada semana con `df -h /` y `du -xsh /root/* | sort -h | tail`; `go clean -cache` es seguro. Un disco lleno mata al nodo igual que el OOM del 21-sep.
- **RAM:** no correr `go test ./...` entero con el nodo arriba sin `-p 1`; el fuzzing nocturno va con `GOMAXPROCS=2 nice -n 19`.
- **Actualizar:** sección "Actualizar el binario", siempre con `sha256sum -c` y copia de los binarios anteriores en `/root/bin-anterior/<version>/`.

## Limpieza de disco del seed 1 (7-oct-2026)

Encontrado al escribir la receta del seed: disco al **82 %** (59/75 GB) con la cadena viva en 7.4 GB. Se identificó cada cosa antes de tocarla y se borró solo por nombre completo, después de copiar lo que se conserva:

- Journal de systemd: 3.9 GB → 500 MB (`journalctl --vacuum-size=500M`). Caché de Go: 1.2 GB (`go clean -cache`; se regenera).
- Borradas las bases de datos de tres testnets muertas (génesis que ya no existe; su historia está en `TESTNET-RELANZAMIENTOS.md`): `rupix-testnet-OLD-0.12.22` (6.2 GB, testnet privada anterior al relanzamiento #1), `rupix-testnet-OLD-halving10k-1788491577` (4.2 GB, testnet #1), `rupix-testnet-v052-vieja` (6.2 GB, testnet #3).
- Borrada la devnet de la "fábrica" de la escalera de agosto (`rupix-escalera/rupix-devnet`, 8.5 GB) y su copia vieja; **conservados** en `/root/retirados/escalera-ago2026/`: `fabrica.sh`, sus logs, `node.log`, `miner.log` y la wallet de devnet (`wallet/keys.json`).
- Borrado `/root/repro` (residuo de `verificar-binarios.sh`; se regenera).
- **No se tocó:** `rupixd-v3` (769 MB: el código de antes del 18-jun-2026, la era descartada, que no está en el repo; es historia), las devnets de septiembre (`rupix-devnet-algo`, `rupix-devnet-v06`), `/root/go`, `/root/bin-anterior/`.

Resultado: **41 %** (30/75 GB). Verificado también que `seed.rupix.network` resuelve a la IP del seed (el `DNSSeeds` de `params.go` funciona; un nodo sin `--addpeer` encuentra la red). Regla desde hoy: `df -h /` en cada cierre de semana; nada por encima del 70 % sin explicación.

