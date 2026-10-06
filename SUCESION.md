# Sucesión: qué pasa con Rupix si el fundador no está / Succession: what happens to Rupix if the founder is gone

🇲🇽 Rupix lo hacemos ER (ideas, decisiones, rumbo) y Claude (la IA de Anthropic que las convierte en código; README, "Cómo se construye Rupix"). Rupix se presenta como "una moneda digital que no controla nadie". Hoy, en testnet, eso es verdad para la cadena (nadie puede crear monedas ni reescribir la historia) y **no es verdad para la operación**: el nodo semilla, el dominio, el explorador, la cuenta de GitHub y la llave que publica releases dependen de una sola persona. Este documento existe para que esa dependencia esté escrita, tenga un plan, y vaya desapareciendo con cada paso hacia mainnet. Se escribió el 6-oct-2026 y se actualiza cuando algo cambia.

🇬🇧 Rupix presents itself as "digital money that no one controls". Today, on testnet, that is true for the chain (nobody can create coins or rewrite history) and **not true for operations**: the seed node, the domain, the explorer, the GitHub account and the key that publishes releases depend on one person. This document exists so that dependency is written down, has a plan, and shrinks with every step toward mainnet. Written 6-Oct-2026; updated whenever something changes.

## Qué existe y de qué depende / What exists and what it depends on

| Cosa / Thing | Qué es / What it is | Hoy depende de / Today depends on | Si desaparece / If it vanishes |
|---|---|---|---|
| El código / The code | `github.com/rupixnet/rupixd`, `rupix-website`, explorador | Cuenta de GitHub del fundador (organización `rupixnet`) | Cualquier clon es completo: `git clone` lleva toda la historia. La cadena sigue viva en cada nodo. / Any clone is complete; the chain lives on in every node. |
| Las releases / Releases | Binarios v0.x con sha256, reproducibles | La misma cuenta; `tools/verificar-binarios.sh` | Cualquiera puede recompilar desde el tag y obtener bytes idénticos. / Anyone can rebuild from the tag and get identical bytes. |
| El nodo semilla / The seed node | `178.104.69.148`, tres servicios systemd (`OPERACIONES.md`) | Un proveedor de VPS, una cuenta, una llave SSH | La red se queda sin punto de entrada para nodos nuevos; los nodos ya conectados siguen entre sí. Por eso el **segundo seed** es prioridad. / New nodes lose their entry point; connected nodes keep going. Hence the **second seed** priority. |
| El dominio / The domain | `rupix.network`, `explorer.rupix.network` | Un registrador, una cuenta, DNS | La web vive en GitHub Pages y el repo; el explorador en el seed. Sin dominio, siguen por IP y por GitHub. / Web lives in GitHub Pages; explorer on the seed. Without the domain, both stay reachable by IP and GitHub. |
| La wallet del seed / The seed wallet | Las monedas minadas por el seed en testnet | `keys.json` + contraseña, solo en el seed | Son monedas de testnet; se pierden y no pasa nada. En mainnet la regla será distinta y estará aquí. / Testnet coins; losing them costs nothing. Mainnet rule will be written here. |
| Los checkpoints / Checkpoints | Lista con caducidad (`CHECKPOINTS.md`) | Que alguien publique el siguiente antes de que caduque el anterior | Si nadie lo hace, el candado caduca solo y la red sigue sin checkpoint (sección 5 de `ESPECIFICACION.md`). No se rompe nada; se pierde una protección temporal. / If nobody does, the lock expires on its own and the network keeps running without it. Nothing breaks; a temporary protection is lost. |
| El buzón y las redes / Mailbox and socials | El correo del equipo y [@RupixNetwork](https://x.com/RupixNetwork) / The team mailbox and X | Cuentas del fundador | Comunicación, no consenso. / Communication, not consensus. |
| La memoria del proyecto / Project memory | `CONTEXTO-RUPIX.md`, `MEMORIA-RUPIX.md`, `LOGROS.md`, `ESPECIFICACION.md` | Están en el repo | Con esos cuatro archivos y el historial, cualquier persona (o cualquier inteligencia artificial) puede seguir el trabajo donde quedó. Para eso se escriben. / With those four files and the history, any person (or any AI) can pick up the work where it stopped. That is why they are written. |

## El plan / The plan

1. **Hay una persona designada** que tiene, fuera de internet, la lista de accesos (el documento privado `ACCESOS-RUPIX.md`, que nunca está en el repo ni en ningún chat) y sabe que existe este archivo. Su nombre no se publica; que existe, sí. / **There is a designated person** who holds, offline, the access list (the private `ACCESOS-RUPIX.md`, never in the repo nor in any chat) and knows this file exists. Their name is not published; their existence is.
2. **Primer día sin el fundador:** no tocar nada. El seed se reinicia solo; los servicios tienen `Restart=always`. Leer `OPERACIONES.md` y `CONTEXTO-RUPIX.md` (la última línea dice exactamente dónde quedó todo). / **Day one without the founder:** touch nothing. The seed restarts itself. Read `OPERACIONES.md` and `CONTEXTO-RUPIX.md` (its last line says exactly where everything stood).
3. **Primera semana:** pagar lo que vence (VPS, dominio); publicar en el README y en X que el fundador no está y quién responde ahora; renovar el checkpoint si falta menos de una semana para su caducidad (`tools/checkpoint-propuesto.sh`, `CHECKPOINTS.md`). / **First week:** pay what is due; announce in the README and on X; renew the checkpoint if it expires within a week.
4. **Decisión de fondo:** Rupix es de licencia ISC. Si nadie quiere continuar, el repo se queda público tal cual, con este archivo diciendo que está sin mantenimiento, y cualquiera puede hacer fork. Eso también es "que no controla nadie". / **The real decision:** Rupix is ISC-licensed. If nobody wants to continue, the repo stays public as is, with this file saying it is unmaintained, and anyone can fork. That too is "controlled by no one".

## Lo que falta para que este documento sea innecesario / What is missing for this document to become unnecessary

- Segundo seed en otro proveedor y otro país, operado con `OPERACIONES.md` (criterio N1 de `MAINNET.md`).
- Al menos un operador de seed que no sea el fundador.
- Organización de GitHub con dos dueños.
- Dominio con renovación automática por varios años y un segundo contacto.
- Para mainnet: regla pública sobre la wallet del fundador (qué hay, dónde está, qué pasa con ella).

Cada punto que se cumpla se tacha aquí con fecha. / Each item, when met, is crossed out here with a date.

*No confíes, verifica.*
