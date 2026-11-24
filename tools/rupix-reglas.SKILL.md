---
name: "rupix-reglas"
description: "Reglas permanentes y procedimiento de trabajo de Rupix (rupixd, web, seed). Úsala en cualquier tarea de Rupix: cambios de código, docs, releases, bundles, checkpoints, web."
---

# Rupix: reglas permanentes

Rupix es un fork de kaspad en Go (42M de techo, cero premine, escalera de gemas Gold/Diamante/Platino/Rodio/Kings = versiones de script 0..4). Lo hacen ER y Claude. Frase pública: «Rupix está creado por ER con Claude, la inteligencia artificial de Anthropic: todo el código, las pruebas y los documentos están hechos con Claude.» Firma: «No confíes, verifica. ER y Claude».

## Lo permanente (nunca se negocia)
1. Todo honesto y verificable. Si no se verificó, se dice «no lo he verificado». Nada se publica con una sola fuente (ubicaciones, cifras).
2. Todo actualizado el mismo día y en los dos idiomas (ES/EN): repo, web, explorador, guías, CONTEXTO-RUPIX.md, LOGROS, MEMORIA, CHANGELOG, ESPECIFICACION, MAINNET. «Nada suelto.»
3. Nada viejo se borra: va a `retirados/` o a MEMORIA. Commits publicados nunca se reescriben.
4. Sin secretos en repo, docs ni chat: ni contraseñas, ni emails, ni la clave de la wallet del seed, ni `ACCESOS-RUPIX.md`. Nunca publicar ni guardar la IP de un usuario; nunca investigar quién es un nodo.
5. Tests primero rojos, luego verdes. Toda sustitución de texto hace `assert` de su objetivo. `set -o pipefail` siempre.
6. Nunca congelar ni revertir fondos, «ni para bien»: un interruptor de congelación es centralización.
7. Nada se regala ni se paga por participar; reconocimiento solo público y verificable en la web; nunca pedir seeds ni dinero. Nada se instala solo en la PC de nadie; toda herramienta que instala muestra el plan y pide «si».
8. Mainnet: criterios, no fecha (MAINNET.md). Dos meses de aviso. MAINNET.md queda BORRADOR hasta que ER lo lea. Antes de lanzar algo público: parar y hablarlo con ER.
9. Siempre mejorar seguridad y experiencia del usuario; guardar todo; probar antes de decir o hacer.

## Cómo se mueve el código
- El contenedor de Claude NO puede hacer push (proxy 403) ni compilar (Go 1.24; el proyecto pide 1.26). Claude arma un **bundle con nombre único** (`git bundle create`, `git bundle verify`, sha256) y lo manda por chat; ER lo sube con `scp` al seed (`ssh root@178.104.69.148`, Go 1.26.6) y corre allí **una sola cadena `&&`** que: verifica sha256 → `git fetch` del bundle a una rama con nombre distinto (nunca a la rama checked-out) → tests → push desde la rama (`git push origin rama:main`) → `echo "PUSH OK: $(git rev-parse --short HEAD)"`.
- Claude verifica después en GitHub con `git fetch origin main:refs/remotes/origin/main` y mira el hash. Sin eso, no está hecho.
- Trabajo en el seed: worktree `/root/rupixd-rel`; `/root/rupixd` no se toca mientras corre un fuzz.
- Nunca esperar procesos con `pgrep -f` y una cadena que el propio esperador contiene; usar PID o archivo. Nada se deja «encadenado» sin comprobar al día siguiente que arrancó.
- Avisos del stop-hook de «unpushed/uncommitted»: la respuesta es siempre «no» (lo empuja ER desde el seed).

## Releases
- Ramas `rel-vX.Y.Z` (nunca iguales al tag); tags como `refs/tags/vX.Y.Z`.
- La suite de tests y la publicación van en la MISMA cadena `&&` (lección de v0.6.8). `gh release create --notes-file releases/vX.Y.Z.md --verify-tag`; CI ~10 min; runners fijos (`ubuntu-24.04, windows-2022, macos-14`); `tools/verificar-binarios.sh` exige Go 1.26.6 y gcc 13.x; 4/4 IGUAL o no se publica. Un release que da DIFIERE se marca «no usar» y se reemplaza; nunca se despliega «mientras tanto».
- Después: actualizar seed por zip + `SHA256SUMS-linux.txt`; grep de la versión vieja en repo y web; instaladores `ULTIMA_CONOCIDA` + sha256 en las 4 guías; CHANGELOG, releases/vX.Y.Z.md ES/EN, hito en la web.

## Checkpoints
- Temporales, con caducidad (`CheckpointsExpireDAAScore`). Se proponen con `tools/checkpoint-propuesto.sh <pruningPointHash externo>` y se confirman con un nodo externo (JC: `GetBlock <hash> false`) ANTES de publicarse. Se anotan en CHECKPOINTS ES/EN y LOGROS.

## Herramientas de verificación
- Toda herramienta distingue «no comprobable» (AVISO) de «mal» (FALLO).
- Comandos para «la máquina nueva» empiezan con `test "$(hostname)" != rupix-seed-1 &&`.
- El daemon de la wallet del seed queda parado salvo para usarlo (`systemctl stop rupixwallet-testnet` al final).

## Fuentes de verdad en el repo
`CONTEXTO-RUPIX.md` (diario), `MEMORIA.md` (lecciones; «Lo permanente» arriba), `OPERACIONES.md` (procedimientos), `MAINNET.md` (criterios), `CHECKPOINTS.md`, `ESPECIFICACION.md`. Antes de responder sobre el estado del proyecto, leerlos; si ER recuerda algo distinto de lo que dice un archivo, verificar antes de contestar.