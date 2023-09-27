package util

import (
	"bytes"
	"encoding/hex"
	"log"
	"math/big"
	"testing"
)

func TestEncodePrefixVarint(t *testing.T) {
	type testCase struct {
		value    *big.Int
		expected string
	}

	testCases := []testCase{
		{value: big.NewInt(0), expected: "00"},
		{value: big.NewInt(1), expected: "01"},
		{value: big.NewInt(127), expected: "7f"},
		{value: big.NewInt(128), expected: "8002"},
		{value: big.NewInt(16383), expected: "bfff"},
		{value: big.NewInt(16384), expected: "c00002"},
		{value: big.NewInt(2097151), expected: "dfffff"},
		{value: big.NewInt(2097152), expected: "e0000002"},
		{value: big.NewInt(21000000), expected: "e0f40614"},
		{value: big.NewInt(268435455), expected: "efffffff"},
		{value: big.NewInt(268435456), expected: "f000000002"},
		{value: big.NewInt(34359738367), expected: "f7ffffffff"},
		{value: big.NewInt(34359738368), expected: "f80000000002"},
		{value: big.NewInt(4398046511103), expected: "fbffffffffff"},
		{value: big.NewInt(4398046511104), expected: "fc000000000002"},
		{value: big.NewInt(562949953421311), expected: "fdffffffffffff"},
		{value: big.NewInt(562949953421312), expected: "fe00000000000002"},
		{value: big.NewInt(72057594037927935), expected: "feffffffffffffff"},
		{value: big.NewInt(72057594037927936), expected: "ff0000000000000001"},
	}

	for _, tc := range testCases {
		buf := bytes.NewBuffer([]byte{})
		iout := EncodePrefixVarint(tc.value, buf)
		hexOut := hex.EncodeToString(buf.Bytes()[:iout])
		log.Printf("%v %v %v", hexOut, iout, buf.Bytes())

		if tc.expected != hexOut {
			t.Fatalf("Expected %s - Got %s", tc.expected, hexOut)
		}
	}
}

func TestDecodePrefixVarint(t *testing.T) {
	buf := bytes.NewBuffer([]byte{255, 152, 120, 6, 1, 0, 0, 0, 0, 18})
	i, err := DecodePrefixVarint(buf)
	if err != nil {
		t.Fatal(err)
	}

	log.Println("I", i)
}
