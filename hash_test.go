package proof

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/ethereum/go-ethereum/rlp"
)

func TestPreimage(t *testing.T) {
	cases := map[string]struct {
		node *shortNode
	}{
		"val": {node: shortNodeHex(t, "04040505040804090505040804090505040810", "43445548495548495548")},
		"ref": {node: shortNodeRefHex(t, "04040505040804090505040804090505040810", "43445548495548495548")},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Logf("Orginal: %s", tc.node)
			bz, err := rlp.EncodeToBytes(collapseShortNode(tc.node))
			if err != nil {
				t.Fatalf("Encoding: %+v", err)
			}
			t.Logf("Encoded: %X", bz)
			decoded, err := decodeNode(nil, bz, 0)
			if err != nil {
				t.Fatalf("Decoding: %+v", err)
			}
			t.Logf("Parsed: %s", decoded)
		})
	}

}

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
		// TestEthTrie/ends_with_value_node
		// proof_test.go:95: -> C43EE591D2EF0F00C6CD919F545E74E3AF65E0B9C24971D9B2C5A85D69FC3A7C
		// proof_test.go:96: ---> (4) [
		//       0: <nil> 1: <nil> 2: <nil> 3: <nil> 4: <40ce5b8ec04c2b4b135ed7f0c8e42f36af377f7c75b4fa3283e8ab2a95d8a88f> 5: <nil> 6: [
		//         0: <nil> 1: {10: 61 } 2: {10: 62 } 3: <nil> 4: <nil> 5: <nil> 6: <nil> 7: <nil> 8: <nil> 9: <nil> a: <nil> b: <nil> c: <nil> d: <nil> e: <nil> f: <nil> [17]: <nil>
		//       ] 7: <nil> 8: <nil> 9: <nil> a: <nil> b: <nil> c: <nil> d: <nil> e: <nil> f: <nil> [17]: <nil>
		//     ]
		"inner full node": {
			node: sparseFullNode(kids{
				4: hashNode(fromHex(t, "40ce5b8ec04c2b4b135ed7f0c8e42f36af377f7c75b4fa3283e8ab2a95d8a88f")),
				6: sparseFullNode(kids{1: shortNodeValue("", "a"), 2: shortNodeValue("", "b")}),
			}),
			expect: fromHex(t, "C43EE591D2EF0F00C6CD919F545E74E3AF65E0B9C24971D9B2C5A85D69FC3A7C"),
		},
		// proof_test.go:95: -> 40CE5B8EC04C2B4B135ED7F0C8E42F36AF377F7C75B4FA3283E8AB2A95D8A88F
		// proof_test.go:96: ---> (1) [
		//       0: <nil> 1: {10: 41 } 2: <27f91a1c3a6df630ec54febc2c39dbaac65f4985d3ffa4c0dd78bf870c95c96d> 3: {04040505040804090505040804090505040810: 43445548495548495548 } 4: {040a040f0409040f040904080406050710: 444a4f494f49484657 } 5: {04080406040b040804050408040f04080507040f0408040610: 4548464b4845484f48574f4846 } 6: <nil> 7: <nil> 8: <nil> 9: <nil> a: <nil> b: <nil> c: <nil> d: <nil> e: <nil> f: <nil> [17]: <nil>
		//     ]
		"inner full node 2": {
			node: sparseFullNode(kids{
				1: shortNodeValue("", "A"),
				2: hashNode(fromHex(t, "27f91a1c3a6df630ec54febc2c39dbaac65f4985d3ffa4c0dd78bf870c95c96d")),
				3: shortNodeHex(t, "04040505040804090505040804090505040810", "43445548495548495548"),
				4: shortNodeHex(t, "040a040f0409040f040904080406050710", "444a4f494f49484657"),
				5: shortNodeHex(t, "04080406040b040804050408040f04080507040f0408040610", "4548464b4845484f48574f4846"),
			}),
			expect: fromHex(t, "40CE5B8EC04C2B4B135ED7F0C8E42F36AF377F7C75B4FA3283E8AB2A95D8A88F"),
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

func shortNodeHex(t *testing.T, hkey, hval string) *shortNode {
	return &shortNode{
		Key: fromHex(t, hkey),
		Val: valueNode(fromHex(t, hval)),
	}
}

func shortNodeRefHex(t *testing.T, hkey, hval string) *shortNode {
	return &shortNode{
		Key: fromHex(t, hkey),
		Val: hashNode(fromHex(t, hval)),
	}
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
