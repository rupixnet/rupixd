package constants

// Niveles de Rupix — la escalera de escasez.
// Cada nivel se codifica en ScriptPublicKey.Version del output:
//
//	0 = Gold (divisible en rupias, minable, paga fees y burns)
//	1..4 = gemas: piezas enteras e indivisibles (monto siempre = 1)
//
// Crear 1 gema de nivel N exige quemar exactamente BurnRatio del nivel N-1,
// y solo a partir del halving que la desbloquea.
const (
	LevelGold     uint16 = 0
	LevelDiamante uint16 = 1
	LevelPlatino  uint16 = 2
	LevelRodio    uint16 = 3
	LevelKings    uint16 = 4

	// BurnRatio: cuantas piezas del nivel inferior se destruyen por 1 del superior
	BurnRatio = 10

	// GemAmount: el monto de todo output de gema. Las gemas no se fraccionan.
	GemAmount = 1

	// MaxKings: tope absoluto del nivel mas alto. El King 2,101 no puede existir.
	MaxKings = 2_100

	// Burn por transaccion: cada tx (salvo la coinbase) destruye Gold por
	// protocolo — no es fee, no va al minero, no va a nadie. Anti-spam de raiz:
	// la base castiga el volumen, el por-byte castiga el peso.
	BurnBase    = uint64(1_000) // rupias fijas por transaccion
	BurnPerByte = uint64(10)    // rupias por cada byte de la transaccion
)

// LevelUnlockDaaScore devuelve el DAA score desde el cual un nivel puede crearse.
// Cada gema nace en su halving: Diamante en el 1, Platino en el 2,
// Rodio en el 3, Kings en el 4. Gold existe desde el genesis.
func LevelUnlockDaaScore(level uint16, blocksPerHalving uint64) uint64 {
	if level == LevelGold {
		return 0
	}
	return uint64(level) * blocksPerHalving
}
