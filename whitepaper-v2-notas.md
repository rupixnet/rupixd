
## DEFINICIÓN HONESTA DE "PARA TODOS" (reflexión de Edu)
"Para todos" NO = gratis/sin esfuerzo. SÍ = sin barreras de privilegio.
- No necesitas permiso, ni ser rico, ni empresa. No hay club cerrado.
- Con inversión pequeña y al alcance (PC gamer $500-1500 o rentar GPU $/día), cualquiera entra.
- Vs ASICs (decenas de miles), es un salto enorme en accesibilidad.
Analogía: "cualquiera puede tener un changarro" — no gratis, pero accesible sin privilegios.
Los ASICs aparecen si Rupix triunfa (señal de valor, problema de éxito). Para entonces:
comunidad + opción de migrar a Autolykos para proteger el "para todos".
MENSAJE HONESTO: "Para todos: cualquiera con inversión pequeña y al alcance mina, sin
permisos ni privilegios." Realista, justo, sostenible. No prometer "gratis" (mentira/insostenible).

## PÚBLICO REAL DE RUPIX (observación clave de Edu)
Quien mina cripto YA es gente tech/cripto que a menudo YA tiene GPU o la consigue fácil.
"El que sabe de cripto, entiende Rupix y le interesa, sin pensarlo minaría/compraría."
→ GPU (RupixHeavyHash) es PERFECTO para ese público. No hace falta CPU puro (Autolykos).
"Para todos" real = comunidad global cripto/tech (México, India, Argentina...), sin
privilegios, con lo que ya tienen o inversión chica. NO literal "8 mil millones minan".
CONCLUSIÓN FINAL DEL ALGORITMO: RupixHeavyHash (GPU, fácil, 3 archivos) es el camino
correcto — no un compromiso, sino LO ideal para el público real. Autolykos = plan B si
aparecen ASICs (señal de éxito). La pieza más difícil se volvió un plan claro y manejable.

## RESUMEN DEL DÍA (cierre) — todo visible en repo
CERRADO: H-4, H-6 (Kings en proof, el grave), SHA256, textos Kaspa→Rupix, quema explicada en web, pruning mergeado a main.
VIVO: testnet renacida (halving 100k), Coco mina Gold real, PRIMERA TRANSFERENCIA (tx 07c6b1ab...) + quema.
PENDIENTE AUDITOR: cero mentiroso (multi-peer, complejo), H-1 (mapeado), verificable total (commitment header), 22 vulns deps.
IDEAS GUARDADAS: verificador de tx, Muro de Fundadores, contador de nodos, UX wallet gema, blog (Coco), Cerebro de Rupix.
DONDE VAMOS: fin Etapa 2 + corazón de Etapa 3 hecho. ~40% a mainnet. Lo más difícil conceptual (pruning) cruzado.
PARA X (cierre del día): resumen de avances + tx histórica 07c6b1ab...

## ALGORITMO — MAPA COMPLETO (radiografias hechas, listo para cirugia)
ARCHIVOS: pow.go (112) + heavyhash.go (91) + xoshiro.go (38). Chicos, aislados.

EL CORAZON (heavyhash.go):
- Linea 11: matrix [64][64]uint16 (la matriz)
- Linea 15: newxoShiRo256PlusPlus(hash) <- EL PRNG a reemplazar
- Linea 25: computeRank()==64 (NO TOCAR - trampa auditor)

EL PRNG (xoshiro.go) - lo que cambiamos:
- Estado s0,s1,s2,s3 (del hash del bloque)
- Uint64(): res=rotl(s0+s3,23)+s0; t=s1<<17; s2^=s0; s3^=s1; s1^=s2; s0^=s3; s2^=t; s3=rotl(s3,45)
- ESTO tienen los ASICs de Kaspa en silicio.

PLAN CIRUGIA RupixHeavyHash (cambio ESTRUCTURAL del PRNG):
1. Crear rupixprng.go (mismo estado del hash, OTRA formula estructural)
2. heavyhash.go linea 15: newxoShiRo256PlusPlus -> newRupixPRNG
3. Vectores de prueba nuevos (xoshiro_test tiene los viejos)
4. Probar mina/valida/rango64
5. Relanzar testnet (hash distinto, como el commitment)

DECISION PENDIENTE: que formula para RupixPRNG.
- Constante sola = DEBIL (auditor). Cambio ESTRUCTURAL = otro algoritmo/mas operaciones.
- Candidatos: variante con operaciones extra, PCG64, SplitMix, o xoshiro+capa extra.
- Es EL corazon de la seguridad -> pensar con calma.

TRAMPAS AUDITOR: (1) rango 64 (2) no desbordar uint16 (3) no tocar computeRank (4) vectores nuevos.
