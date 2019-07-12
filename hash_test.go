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
			node:   &shortNode{Key: toKey("1"), Val: valueNode("1")},
			expect: fromHex(t, "0FF8F6CCAB8202455F6C51C66FB2436B3009D0E3BA58225CFDAEE4CC973D8FF2"),
		},
		"longer value": {
			node:   &shortNode{Key: toKey("fooled"), Val: valueNode("fooled")},
			expect: fromHex(t, "F4CE6E6AE7FE32D3748FD8BBD8B7CE8355992C328757E99C4BB81E3BEA989D88"),
		},
		"longest value": {
			node:   &shortNode{Key: toKey("more than 16 bytes here..."), Val: valueNode("more than 16 bytes here...")},
			expect: fromHex(t, "E66333E75B4D83C31F0EDE8B9FF73455EA656D6C95A74D6134027F3F2E8803EE"),
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

func toKey(key string) []byte {
	return keybytesToHex([]byte(key))
}

func fromHex(t testing.TB, hexstr string) []byte {
	res, err := hex.DecodeString(hexstr)
	if err != nil {
		t.Fatalf("Cannot decode hex: %s", hexstr)
	}
	return res
}
