// Copyright (c) 2013, 2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package util_test

import (
	"github.com/rupixnet/rupixd/domain/consensus/utils/constants"
	"math"
	"testing"

	. "github.com/rupixnet/rupixd/util"
)

func TestAmountCreation(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		valid    bool
		expected Amount
	}{
		// Positive tests.
		{
			name:     "zero",
			amount:   0,
			valid:    true,
			expected: 0,
		},
		{
			name:     "max producible",
			amount:   29e9,
			valid:    true,
			expected: Amount(constants.MaxRupia),
		},
		{
			name:     "one hundred",
			amount:   100,
			valid:    true,
			expected: 100 * constants.RupiaPerRupix,
		},
		{
			name:     "fraction",
			amount:   0.01234567,
			valid:    true,
			expected: 1234567,
		},
		{
			name:     "rounding up",
			amount:   54.999999999999943157,
			valid:    true,
			expected: 55 * constants.RupiaPerRupix,
		},
		{
			name:     "rounding down",
			amount:   55.000000000000056843,
			valid:    true,
			expected: 55 * constants.RupiaPerRupix,
		},

		// Negative tests.
		{
			name:   "not-a-number",
			amount: math.NaN(),
			valid:  false,
		},
		{
			name:   "-infinity",
			amount: math.Inf(-1),
			valid:  false,
		},
		{
			name:   "+infinity",
			amount: math.Inf(1),
			valid:  false,
		},
	}

	for _, test := range tests {
		a, err := NewAmount(test.amount)
		switch {
		case test.valid && err != nil:
			t.Errorf("%v: Positive test Amount creation failed with: %v", test.name, err)
			continue
		case !test.valid && err == nil:
			t.Errorf("%v: Negative test Amount creation succeeded (value %v) when should fail", test.name, a)
			continue
		}

		if a != test.expected {
			t.Errorf("%v: Created amount %v does not match expected %v", test.name, a, test.expected)
			continue
		}
	}
}

func TestAmountUnitConversions(t *testing.T) {
	tests := []struct {
		name      string
		amount    Amount
		unit      AmountUnit
		converted float64
		s         string
	}{
		{
			name:      "MRUPIX",
			amount:    Amount(constants.MaxRupia),
			unit:      AmountMegaRUPIX,
			converted: 29000,
			s:         "29000 MRUPIX",
		},
		{
			name:      "kRUPIX",
			amount:    44433322211100,
			unit:      AmountKiloRUPIX,
			converted: 444.33322211100,
			s:         "444.333222111 kRUPIX",
		},
		{
			name:      "RUPIX",
			amount:    44433322211100,
			unit:      AmountRUPIX,
			converted: 444333.22211100,
			s:         "444333.222111 RUPIX",
		},
		{
			name:      "mRUPIX",
			amount:    44433322211100,
			unit:      AmountMilliRUPIX,
			converted: 444333222.11100,
			s:         "444333222.111 mRUPIX",
		},
		{

			name:      "μRUPIX",
			amount:    44433322211100,
			unit:      AmountMicroRUPIX,
			converted: 444333222111.00,
			s:         "444333222111 μRUPIX",
		},
		{

			name:      "rupia",
			amount:    44433322211100,
			unit:      Amountrupia,
			converted: 44433322211100,
			s:         "44433322211100 rupia",
		},
		{

			name:      "non-standard unit",
			amount:    44433322211100,
			unit:      AmountUnit(-1),
			converted: 4443332.2211100,
			s:         "4443332.22111 1e-1 RUPIX",
		},
	}

	for _, test := range tests {
		f := test.amount.ToUnit(test.unit)
		if f != test.converted {
			t.Errorf("%v: converted value %v does not match expected %v", test.name, f, test.converted)
			continue
		}

		s := test.amount.Format(test.unit)
		if s != test.s {
			t.Errorf("%v: format '%v' does not match expected '%v'", test.name, s, test.s)
			continue
		}

		// Verify that Amount.ToRUPIX works as advertised.
		f1 := test.amount.ToUnit(AmountRUPIX)
		f2 := test.amount.ToRUPIX()
		if f1 != f2 {
			t.Errorf("%v: ToRUPIX does not match ToUnit(AmountRUPIX): %v != %v", test.name, f1, f2)
		}

		// Verify that Amount.String works as advertised.
		s1 := test.amount.Format(AmountRUPIX)
		s2 := test.amount.String()
		if s1 != s2 {
			t.Errorf("%v: String does not match Format(AmountRUPIX): %v != %v", test.name, s1, s2)
		}
	}
}

func TestAmountMulF64(t *testing.T) {
	tests := []struct {
		name string
		amt  Amount
		mul  float64
		res  Amount
	}{
		{
			name: "Multiply 0.1 RUPIX by 2",
			amt:  100e5, // 0.1 RUPIX
			mul:  2,
			res:  200e5, // 0.2 RUPIX
		},
		{
			name: "Multiply 0.2 RUPIX by 0.02",
			amt:  200e5, // 0.2 RUPIX
			mul:  1.02,
			res:  204e5, // 0.204 RUPIX
		},
		{
			name: "Round down",
			amt:  49, // 49 rupia
			mul:  0.01,
			res:  0,
		},
		{
			name: "Round up",
			amt:  50, // 50 rupia
			mul:  0.01,
			res:  1, // 1 rupia
		},
		{
			name: "Multiply by 0.",
			amt:  1e8, // 1 RUPIX
			mul:  0,
			res:  0, // 0 RUPIX
		},
		{
			name: "Multiply 1 by 0.5.",
			amt:  1, // 1 rupia
			mul:  0.5,
			res:  1, // 1 rupia
		},
		{
			name: "Multiply 100 by 66%.",
			amt:  100, // 100 rupia
			mul:  0.66,
			res:  66, // 66 rupia
		},
		{
			name: "Multiply 100 by 66.6%.",
			amt:  100, // 100 rupia
			mul:  0.666,
			res:  67, // 67 rupia
		},
		{
			name: "Multiply 100 by 2/3.",
			amt:  100, // 100 rupia
			mul:  2.0 / 3,
			res:  67, // 67 rupia
		},
	}

	for _, test := range tests {
		a := test.amt.MulF64(test.mul)
		if a != test.res {
			t.Errorf("%v: expected %v got %v", test.name, test.res, a)
		}
	}
}



