package burnmanager

import (
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/ruleerrors"
	"github.com/pkg/errors"
)

// Niveles de Rupix â€” sistema de escasez progresiva
const (
	LevelL1 = uint8(1) // Rupix Oro      ðŸ¥‡ â€” minado
	LevelL2 = uint8(2) // Rupix Diamante ðŸ’Ž â€” quemar 10 L1
	LevelL3 = uint8(3) // Rupix Platino  â¬œ â€” quemar 10 L2
	LevelL4 = uint8(4) // Rupix Rodio    ðŸ”´ â€” quemar 10 L3
	LevelL5 = uint8(5) // Kings Rupix    ðŸ‘‘ â€” quemar 10 L4
)

// Supply maximo por nivel â€” matematica irreversible
const (
	MaxSupplyL1 = uint64(42_000_000)
	MaxSupplyL2 = uint64(2_100_000)
	MaxSupplyL3 = uint64(210_000)
	MaxSupplyL4 = uint64(21_000)
	MaxSupplyL5 = uint64(2_100)
)

// BurnRatio cuantos tokens del nivel anterior se queman para obtener 1 del siguiente
const BurnRatio = uint64(10)

// BurnAddress es la direccion de quema â€” sin llave privada, nadie puede recuperar
const BurnAddress = "rupix:qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq"

// BurnManager maneja la logica de quema y niveles
type BurnManager struct{}

// New crea una nueva instancia del BurnManager
func New() *BurnManager {
	return &BurnManager{}
}

// ValidateBurnTransaction valida que una TX de quema sea correcta
func (bm *BurnManager) ValidateBurnTransaction(tx *externalapi.DomainTransaction) error {
	// Verificar que tenga outputs
	if len(tx.Outputs) == 0 {
		return errors.Wrapf(ruleerrors.ErrNoTxInputs, "burn transaction has no outputs")
	}

	// Verificar que el payload tenga el nivel solicitado
	if len(tx.Payload) < 2 {
		return errors.Wrapf(ruleerrors.ErrInvalidPayload,
			"burn transaction payload too small: need level and source level")
	}

	sourceLevel := tx.Payload[0] // nivel origen
	targetLevel := tx.Payload[1] // nivel destino

	// Validar niveles
	if err := validateLevelTransition(sourceLevel, targetLevel); err != nil {
		return err
	}

	// Validar monto quemado
	var totalBurned uint64
	for _, out := range tx.Outputs {
		totalBurned += out.Value
	}

	requiredBurn := BurnRatio // 10 tokens del nivel anterior
	if totalBurned < requiredBurn {
		return errors.Wrapf(ruleerrors.ErrInvalidPayload,
			"burn amount %d is less than required %d for level transition %d->%d",
			totalBurned, requiredBurn, sourceLevel, targetLevel)
	}

	return nil
}

// validateLevelTransition verifica que la transicion de nivel sea valida
func validateLevelTransition(sourceLevel, targetLevel uint8) error {
	// Solo se puede subir un nivel a la vez
	if targetLevel != sourceLevel+1 {
		return errors.Wrapf(ruleerrors.ErrInvalidPayload,
			"invalid level transition: can only go from level %d to %d, got %d",
			sourceLevel, sourceLevel+1, targetLevel)
	}

	// Nivel maximo es L5
	if targetLevel > LevelL5 {
		return errors.Wrapf(ruleerrors.ErrInvalidPayload,
			"invalid target level %d: maximum level is %d", targetLevel, LevelL5)
	}

	// Nivel minimo origen es L1
	if sourceLevel < LevelL1 {
		return errors.Wrapf(ruleerrors.ErrInvalidPayload,
			"invalid source level %d: minimum level is %d", sourceLevel, LevelL1)
	}

	return nil
}

// MaxSupplyForLevel devuelve el supply maximo para un nivel dado
func MaxSupplyForLevel(level uint8) (uint64, error) {
	switch level {
	case LevelL1:
		return MaxSupplyL1, nil
	case LevelL2:
		return MaxSupplyL2, nil
	case LevelL3:
		return MaxSupplyL3, nil
	case LevelL4:
		return MaxSupplyL4, nil
	case LevelL5:
		return MaxSupplyL5, nil
	default:
		return 0, errors.Wrapf(ruleerrors.ErrInvalidPayload,
			"invalid level %d", level)
	}
}

// LevelName devuelve el nombre del nivel
func LevelName(level uint8) string {
	switch level {
	case LevelL1:
		return "Rupix Gold L1"
	case LevelL2:
		return "Rupix Diamante L2"
	case LevelL3:
		return "Rupix Platino L3"
	case LevelL4:
		return "Rupix Rodio L4"
	case LevelL5:
		return "Kings Rupix L5"
	default:
		return "Nivel Invalido"
	}
}
