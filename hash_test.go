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
			node:   shortNodeValue("1", "1"),
			expect: fromHex(t, "0FF8F6CCAB8202455F6C51C66FB2436B3009D0E3BA58225CFDAEE4CC973D8FF2"),
		},
		"longer value": {
			node:   shortNodeValue("fooled", "fooled"),
			expect: fromHex(t, "F4CE6E6AE7FE32D3748FD8BBD8B7CE8355992C328757E99C4BB81E3BEA989D88"),
		},
		"longest value": {
			node:   shortNodeValue("more than 16 bytes here...", "more than 16 bytes here..."),
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

func TestHashFullNode(t *testing.T) {
	// six := &shortNode{
	// 	Key: fromHex(t, "060f060f04020401050210"),
	// 	Val: valueNode(fromHex(t, "666f6f424152")),
	// }

	cases := map[string]struct {
		node   *fullNode
		expect []byte
	}{
		// From: go test -v . -run TestEthTrie/two_levels
		// proof_test.go:95: -> D9BF67E4B6DD1D92FF614D619D14E6D2AE301098FDA55949BE5E1BDAD26383EE
		// proof_test.go:96: ---> (6) [
		//       0: <nil> 1: {10: 61 } 2: <nil> 3: <nil> 4: <nil> 5: <nil> 6: {060f060f04020401050210: 666f6f424152 } 7: <nil> 8: <nil> 9: <nil> a: <nil> b: <nil> c: <nil> d: <nil> e: <nil> f: <nil> [17]: <nil>
		//     ]
		"with shortnodes": {
			node:   sparseFullNode(kids{1: shortNodeValue("", "a"), 6: shortNodeValue("ooBAR", "fooBAR")}),
			expect: fromHex(t, "D9BF67E4B6DD1D92FF614D619D14E6D2AE301098FDA55949BE5E1BDAD26383EE"),
		},
		// From: go test -v . -run TestEthTrie/embeded_full_node
		// proof_test.go:95: -> A18F94B41AC5C5E3D4960A90DDEB9E7426FD79783F6217DC20EB0D06BB472246
		// proof_test.go:96: ---> (6) [
		//       0: <nil> 1: <nil> 2: <nil> 3: <nil> 4: [
		//         0: <nil> 1: {10: 41 } 2: {10: 42 } 3: <nil> 4: <nil> 5: <nil> 6: <nil> 7: <nil> 8: <nil> 9: <nil> a: <nil> b: <nil> c: <nil> d: <nil> e: <nil> f: <nil> [17]: <nil>
		//       ] 5: <nil> 6: [
		//         0: <nil> 1: {10: 61 } 2: {10: 62 } 3: <nil> 4: <nil> 5: <nil> 6: <nil> 7: <nil> 8: <nil> 9: <nil> a: <nil> b: <nil> c: <nil> d: <nil> e: <nil> f: <nil> [17]: <nil>
		//       ] 7: <nil> 8: <nil> 9: <nil> a: <nil> b: <nil> c: <nil> d: <nil> e: <nil> f: <nil> [17]: <nil>
		//     ]
		"embedded fullnode": {
			node: sparseFullNode(kids{
				4: sparseFullNode(kids{1: shortNodeValue("", "A"), 2: shortNodeValue("", "B")}),
				6: sparseFullNode(kids{1: shortNodeValue("", "a"), 2: shortNodeValue("", "b")}),
			}),
			expect: fromHex(t, "A18F94B41AC5C5E3D4960A90DDEB9E7426FD79783F6217DC20EB0D06BB472246"),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			hash := hashFullNode(tc.node)
			// hash := ethHashFullNode(tc.node)
			if !bytes.Equal(tc.expect, hash[:]) {
				t.Fatalf("Expected %X\n     Got %X", tc.expect, hash[:])
			}
		})
	}
}

type kids map[int]node

func sparseFullNode(children kids) *fullNode {
	fn := fullNode{}
	for idx, child := range children {
		fn.Children[idx] = child
	}
	return &fn
}

func shortNodeValue(key, value string) *shortNode {
	return &shortNode{Key: toKey(key), Val: valueNode(value)}
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
