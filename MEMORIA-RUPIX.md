# MEMORIA RUPIX

> La historia de Rupix contada en orden: qué pasó, por qué se decidió cada cosa y qué creímos y después cambiamos.
> Cada hecho lleva su fuente (commit, tag o archivo) para que cualquiera lo verifique.
> Es la base de lo que responderá Stevenson Rux, el asistente de Rupix.
>
> **Estado:** en construcción. Los capítulos marcados *(por escribir)* se escriben con calma, uno por uno, con revisión del fundador.
>
> **Sobre las fechas:** van en hora de la Ciudad de México. Los commits guardan la hora en UTC, así que el trabajo de la noche aparece en git con fecha del día siguiente.

## Índice

1. Por qué existe Rupix *(por escribir)*
2. El primer intento y el reinicio del 11 de junio *(por escribir; material: sección HISTORIA del CONTEXTO, commit `dc6e22c`, tag `v0.2.4-archive-experimental`)*
3. El arranque limpio del 18 de junio *(por escribir)*
4. La economía sellada *(por escribir)*
5. El génesis y la primera transacción *(por escribir)*
6. Los primeros de fuera *(por escribir)*
7. La costura: el sello de gemas y los bugs de septiembre *(por escribir)*
8. El auditor y la honestidad radical *(por escribir)*
9. v0.5: checkpoints y tests en verde *(por escribir)*
10. El test del King y la v0.6.0 *(por escribir)*
11. La comunidad de Kaspa *(por escribir)*
12. El 26 de septiembre *(por escribir)*
13. Hacia dónde va *(por escribir)*
14. Lo que creímos y después cambiamos *(empezado, abajo)*
- Apéndice: línea de tiempo *(por escribir)*

---

## 14. Lo que creímos y después cambiamos

Esta sección no se borra ni se maquilla. Cada entrada dice qué se creía, qué resultó cierto y dónde está la prueba.

### El algoritmo: RandomX y Autolykos → RupixHeavyHash
En notas de otra etapa se planeó migrar a RandomX (minado por CPU) y, después, a Autolykos v2. No se hizo. El algoritmo de Rupix es **RupixHeavyHash**: el motor de Kaspa (kHeavyHash) con generador y semilla propios desde v0.5.0 (16-sep), y dominio keccak propio desde v0.6.0. Las notas viejas quedaron marcadas como de otra etapa el 26-sep (commit `65a6fde`).

### Cómo se describe Rupix: "activo digital escaso" → "moneda digital que no controla nadie"
Hasta el 26-sep, el README, la web y el CONTEXTO decían "activo digital escaso, como el oro" y "valor para guardar". Ese día el fundador lo cambió: "no lo somos". En testnet no tiene valor de mercado, y prometer valor sería vender humo. La regla desde entonces: se dice lo que Rupix es (moneda digital, sin dueño, sin premine, 42 millones que nadie puede cambiar, cada vez más escasa) y no para qué se usa ni cuánto vale; el rumbo lo pone la gente. Commits `c9a3719` y `27f864a` (repo) y `a33b753` (web).

### La alarma de descalificados que no podía sonar
El 23-sep el auditor pidió vigilar los bloques descalificados. La alarma se hizo la noche del 25-sep (commit `0a76230`) y caminaba hacia atrás por la cadena seleccionada. Por construcción nunca podía ver un bloque descalificado, porque esos quedan fuera de esa cadena: dijo "OK" 134 veces sin poder sonar. El mensaje de su commit afirmaba que "habría cazado el bug del King el día 18"; era falso. El 26-sep el nodo empezó a registrar cada descalificación y la alarma pasó a leer ese registro; si no ve actividad, dice ERROR en vez de OK (commit `195940d`). Probada con un log falso. Falta la prueba de punta a punta en devnet.

### "El mempool no valida la escalera"
Una nota de trabajo (`NOTAS-H1-MEMPOOL.md`) decía que el mempool solo revisaba anti-spam y formato. Al leer el código el 26-sep resultó falso: al insertar una transacción, el mempool llama a la validación del consenso, que ejecuta `checkLevelRules`. Quedan pendientes un test que lo demuestre y el Nivel B (topes históricos). Commit `8167c10b`.

### El README que se quedó atrás
El 26-sep el README todavía decía que el fix del bug del King "saldría en v0.6.0", cuando ya había salido, y fechaba al tercer nodo externo (16-sep) en v0.6.0, cuando era v0.5.0. Además presentaba el primer Diamante real, el primer forjador externo y el primer Platino (14-sep) sin aclarar que fueron en testnets anteriores, no en la #4. Corregido en el commit `8167c10b`.

### El CHANGELOG sin su versión más importante
La v0.6.0 salió y el CHANGELOG pasó 6 días sin su entrada. Se agregó el 26-sep (commit `90995c9`).

### "Dos forjadores externos"
El CONTEXTO decía que había dos forjadores externos en la historia. Era uno: JC. Corregido el 26-sep (commit `b15cbf1`).

### "22 vulnerabilidades"
Un plan viejo hablaba de 22 vulnerabilidades en las dependencias; el README dice 0 según govulncheck. El 26-sep se agregó al CONTEXTO un pendiente basado en ese plan viejo, sin verificar, y se quitó el mismo día (commit `be765fc`). Queda correr govulncheck otra vez para confirmar el 0.

### La devnet olvidada
La devnet de pruebas de v0.6.0 quedó prendida dos días en el seed, con 1.8 GB de RAM y su puerto P2P abierto a internet. Se apagó el 26-sep y quedó escrita la regla: las devnets solo corren mientras se usan (commit `8d6884f`).

### La falsa alarma del explorer
El 26-sep se creyó que el explorer mostraba nombres viejos de la escalera ("Zafiro", "Esmeralda"). Venía de una copia vieja en el espacio de trabajo del asistente: el explorer real ya decía Diamante, Platino, Rodio y Kings. La lección: verificar contra la fuente viva antes de corregir.

---

*No confíes, verifica.*
