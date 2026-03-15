package bip32

import "github.com/pkg/errors"

// BitcoinMainnetPrivate is the version that is used for
// bitcoin mainnet bip32 private extended keys.
// Ecnodes to xprv in base58.
var BitcoinMainnetPrivate = [4]byte{
	0x04,
	0x88,
	0xad,
	0xe4,
}

// BitcoinMainnetPublic is the version that is used for
// bitcoin mainnet bip32 public extended keys.
// Ecnodes to xpub in base58.
var BitcoinMainnetPublic = [4]byte{
	0x04,
	0x88,
	0xb2,
	0x1e,
}

// RupixMainnetPrivate is the version that is used for
// rupix mainnet bip32 private extended keys.
// Ecnodes to xprv in base58.
var RupixMainnetPrivate = [4]byte{
	0x03,
	0x8f,
	0x2e,
	0xf4,
}

// RupixMainnetPublic is the version that is used for
// rupix mainnet bip32 public extended keys.
// Ecnodes to kpub in base58.
var RupixMainnetPublic = [4]byte{
	0x03,
	0x8f,
	0x33,
	0x2e,
}

// RupixTestnetPrivate is the version that is used for
// rupix testnet bip32 public extended keys.
// Ecnodes to ktrv in base58.
var RupixTestnetPrivate = [4]byte{
	0x03,
	0x90,
	0x9e,
	0x07,
}

// RupixTestnetPublic is the version that is used for
// rupix testnet bip32 public extended keys.
// Ecnodes to ktub in base58.
var RupixTestnetPublic = [4]byte{
	0x03,
	0x90,
	0xa2,
	0x41,
}

// RupixdevnetPrivate is the version that is used for
// rupix devnet bip32 public extended keys.
// Ecnodes to kdrv in base58.
var RupixdevnetPrivate = [4]byte{
	0x03,
	0x8b,
	0x3d,
	0x80,
}

// RupixdevnetPublic is the version that is used for
// rupix devnet bip32 public extended keys.
// Ecnodes to xdub in base58.
var RupixdevnetPublic = [4]byte{
	0x03,
	0x8b,
	0x41,
	0xba,
}

// RupixSimnetPrivate is the version that is used for
// rupix simnet bip32 public extended keys.
// Ecnodes to ksrv in base58.
var RupixSimnetPrivate = [4]byte{
	0x03,
	0x90,
	0x42,
	0x42,
}

// RupixSimnetPublic is the version that is used for
// rupix simnet bip32 public extended keys.
// Ecnodes to xsub in base58.
var RupixSimnetPublic = [4]byte{
	0x03,
	0x90,
	0x46,
	0x7d,
}

func toPublicVersion(version [4]byte) ([4]byte, error) {
	switch version {
	case BitcoinMainnetPrivate:
		return BitcoinMainnetPublic, nil
	case RupixMainnetPrivate:
		return RupixMainnetPublic, nil
	case RupixTestnetPrivate:
		return RupixTestnetPublic, nil
	case RupixdevnetPrivate:
		return RupixdevnetPublic, nil
	case RupixSimnetPrivate:
		return RupixSimnetPublic, nil
	}

	return [4]byte{}, errors.Errorf("unknown version %x", version)
}

func isPrivateVersion(version [4]byte) bool {
	switch version {
	case BitcoinMainnetPrivate:
		return true
	case RupixMainnetPrivate:
		return true
	case RupixTestnetPrivate:
		return true
	case RupixdevnetPrivate:
		return true
	case RupixSimnetPrivate:
		return true
	}

	return false
}

