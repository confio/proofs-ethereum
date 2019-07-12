package proof

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestHashShortNode(t *testing.T) {
	cases := map[string]struct {
		node   *shortNode
		expect []byte
	}{
		"simple value": {
			node:   &shortNode{Key: []byte{3, 1, 0x10}, Val: valueNode{0x31}},
			expect: fromHex(t, "0FF8F6CCAB8202455F6C51C66FB2436B3009D0E3BA58225CFDAEE4CC973D8FF2"),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			hash := hashShortNode(tc.node)
			if !bytes.Equal(tc.expect, hash[:]) {
				t.Fatalf("Expected %X\n     Got %X", tc.expect, hash[:])
			}
		})
	}
}

func fromHex(t testing.TB, hexstr string) []byte {
	res, err := hex.DecodeString(hexstr)
	if err != nil {
		t.Fatalf("Cannot decode hex: %s", hexstr)
	}
	return res
}
