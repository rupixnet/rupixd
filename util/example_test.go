package util_test

import (
	"fmt"
	"github.com/rupixnet/rupixd/util/difficulty"
	"math"
	"math/big"

	"github.com/rupixnet/rupixd/util"
)

func ExampleAmount() {

	a := util.Amount(0)
	fmt.Println("Zero rupia:", a)

	a = util.Amount(1e8)
	fmt.Println("100,000,000 rupia:", a)

	a = util.Amount(1e5)
	fmt.Println("100,000 rupia:", a)
	// Output:
	// Zero rupia: 0 KAS
	// 100,000,000 rupia: 1 KAS
	// 100,000 rupia: 0.001 KAS
}

func ExampleNewAmount() {
	amountOne, err := util.NewAmount(1)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(amountOne) //Output 1

	amountFraction, err := util.NewAmount(0.01234567)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(amountFraction) //Output 2

	amountZero, err := util.NewAmount(0)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(amountZero) //Output 3

	amountNaN, err := util.NewAmount(math.NaN())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(amountNaN) //Output 4

	// Output: 1 KAS
	// 0.01234567 KAS
	// 0 KAS
	// invalid rupix amount
}

func ExampleAmount_unitConversions() {
	amount := util.Amount(44433322211100)

	fmt.Println("rupia to kKAS:", amount.Format(util.AmountKiloKAS))
	fmt.Println("rupia to KAS:", amount)
	fmt.Println("rupia to MilliKAS:", amount.Format(util.AmountMilliKAS))
	fmt.Println("rupia to MicroKAS:", amount.Format(util.AmountMicroKAS))
	fmt.Println("rupia to rupia:", amount.Format(util.Amountrupia))

	// Output:
	// rupia to kKAS: 444.333222111 kKAS
	// rupia to KAS: 444333.222111 KAS
	// rupia to MilliKAS: 444333222.111 mKAS
	// rupia to MicroKAS: 444333222111 μKAS
	// rupia to rupia: 44433322211100 rupia
}

// This example demonstrates how to convert the compact "bits" in a block header
// which represent the target difficulty to a big integer and display it using
// the typical hex notation.
func ExampleCompactToBig() {
	bits := uint32(419465580)
	targetDifficulty := difficulty.CompactToBig(bits)

	// Display it in hex.
	fmt.Printf("%064x\n", targetDifficulty.Bytes())

	// Output:
	// 0000000000000000896c00000000000000000000000000000000000000000000
}

// This example demonstrates how to convert a target difficulty into the compact
// "bits" in a block header which represent that target difficulty .
func ExampleBigToCompact() {
	// Convert the target difficulty from block 300000 in the bitcoin
	// main chain to compact form.
	t := "0000000000000000896c00000000000000000000000000000000000000000000"
	targetDifficulty, success := new(big.Int).SetString(t, 16)
	if !success {
		fmt.Println("invalid target difficulty")
		return
	}
	bits := difficulty.BigToCompact(targetDifficulty)

	fmt.Println(bits)

	// Output:
	// 419465580
}


