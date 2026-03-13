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
	// Zero rupia: 0 RUPIX
	// 100,000,000 rupia: 1 RUPIX
	// 100,000 rupia: 0.001 RUPIX
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

	// Output: 1 RUPIX
	// 0.01234567 RUPIX
	// 0 RUPIX
	// invalid rupix amount
}

func ExampleAmount_unitConversions() {
	amount := util.Amount(44433322211100)

	fmt.Println("rupia to kRUPIX:", amount.Format(util.AmountKiloRUPIX))
	fmt.Println("rupia to RUPIX:", amount)
	fmt.Println("rupia to MilliRUPIX:", amount.Format(util.AmountMilliRUPIX))
	fmt.Println("rupia to MicroRUPIX:", amount.Format(util.AmountMicroRUPIX))
	fmt.Println("rupia to rupia:", amount.Format(util.Amountrupia))

	// Output:
	// rupia to kRUPIX: 444.333222111 kRUPIX
	// rupia to RUPIX: 444333.222111 RUPIX
	// rupia to MilliRUPIX: 444333222.111 mRUPIX
	// rupia to MicroRUPIX: 444333222111 μRUPIX
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


