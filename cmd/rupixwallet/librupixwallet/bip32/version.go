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

// KaspaMainnetPrivate is the version that is used for
// kaspa mainnet bip32 private extended keys.
// Encodes to rprv in base58. (Rupix)
var KaspaMainnetPrivate = [4]byte{
	0xea,
	0xb4,
	0x04,
	0x55,
}

// KaspaMainnetPublic is the version that is used for
// kaspa mainnet bip32 public extended keys.
// Encodes to rpub in base58. (Rupix)
var KaspaMainnetPublic = [4]byte{
	0xea,
	0xb4,
	0xfa,
	0x80,
}

// KaspaTestnetPrivate is the version that is used for
// kaspa testnet bip32 public extended keys.
// Encodes to rtrv in base58. (Rupix)
var KaspaTestnetPrivate = [4]byte{
	0xeb,
	0x07,
	0x2f,
	0x80,
}

// KaspaTestnetPublic is the version that is used for
// kaspa testnet bip32 public extended keys.
// Encodes to rtub in base58. (Rupix)
var KaspaTestnetPublic = [4]byte{
	0xeb,
	0x08,
	0x24,
	0x80,
}

// KaspaDevnetPrivate is the version that is used for
// kaspa devnet bip32 public extended keys.
// Encodes to rdrv in base58. (Rupix)
var KaspaDevnetPrivate = [4]byte{
	0xe9,
	0xcf,
	0x50,
	0x80,
}

// KaspaDevnetPublic is the version that is used for
// kaspa devnet bip32 public extended keys.
// Encodes to rdub in base58. (Rupix)
var KaspaDevnetPublic = [4]byte{
	0xe9,
	0xd0,
	0x45,
	0x55,
}

// KaspaSimnetPrivate is the version that is used for
// kaspa simnet bip32 public extended keys.
// Encodes to rsrv in base58. (Rupix)
var KaspaSimnetPrivate = [4]byte{
	0xea,
	0xf2,
	0x64,
	0x55,
}

// KaspaSimnetPublic is the version that is used for
// kaspa simnet bip32 public extended keys.
// Encodes to rsub in base58. (Rupix)
var KaspaSimnetPublic = [4]byte{
	0xea,
	0xf3,
	0x59,
	0x55,
}

func toPublicVersion(version [4]byte) ([4]byte, error) {
	switch version {
	case BitcoinMainnetPrivate:
		return BitcoinMainnetPublic, nil
	case KaspaMainnetPrivate:
		return KaspaMainnetPublic, nil
	case KaspaTestnetPrivate:
		return KaspaTestnetPublic, nil
	case KaspaDevnetPrivate:
		return KaspaDevnetPublic, nil
	case KaspaSimnetPrivate:
		return KaspaSimnetPublic, nil
	}

	return [4]byte{}, errors.Errorf("unknown version %x", version)
}

func isPrivateVersion(version [4]byte) bool {
	switch version {
	case BitcoinMainnetPrivate:
		return true
	case KaspaMainnetPrivate:
		return true
	case KaspaTestnetPrivate:
		return true
	case KaspaDevnetPrivate:
		return true
	case KaspaSimnetPrivate:
		return true
	}

	return false
}
