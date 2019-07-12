package proof

import (
	"fmt"
	"hash"

	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"
)

func hashAnyNode(n node) []byte {
	switch tn := n.(type) {
	case *fullNode:
		return hashFullNode(tn)
	case *shortNode:
		return hashShortNode(tn)
	case valueNode:
		// Is this good?
		return tn
	case hashNode:
		// Is this good?
		return tn
	default:
		panic("this cannot be")
	}
}

func collapseShortNode(n *shortNode) *shortNode {
	collapsed := n.copy()
	collapsed.Key = hexToCompact(n.Key)
	return collapsed
}

func hashShortNode(n *shortNode) []byte {
	bz, err := rlp.EncodeToBytes(collapseShortNode(n))
	if err != nil {
		panic("encode error: " + err.Error())
	}
	hash := makeHashNode(bz)

	fmt.Printf("ShortNode: %s\n", n)
	fmt.Printf("Encoded: %X\n", bz)
	// Notes from encoding: 1 byte string is encoded without prefix, longer as 0x80 + N where N is length (for > 127???)
	// Node: {030110: 31 }
	// Encoded: C482203131
	// C4 some type info? 82 - 2 byte string / 2031 - compact key / 31  unprefixed one byte string

	// Node: {0606060f060f060c0605060410: 666f6f6c6564 }
	// Encoded: CF8720666F6F6C656486666F6F6C6564
	// CF some type info? 87 - 7 byte string / 20666F6F6C6564 - compact key (0x20 + string) / 86 - 6 byte string / 666F6F6C6564 value

	return hash
}

func collapseFullNode(n *fullNode) *fullNode {
	collapsed := n.copy()
	for i := 0; i < 16; i++ {
		switch child := collapsed.Children[i].(type) {
		case *shortNode:
			collapsed.Children[i] = collapseShortNode(child)
		case *fullNode:
			collapsed.Children[i] = collapseFullNode(child)
			// leave valueNode and hashNode (or reference) untouched
		}
	}
	return collapsed
}

func hashFullNode(n *fullNode) []byte {
	// this is encoding process
	bz, err := rlp.EncodeToBytes(collapseFullNode(n))
	if err != nil {
		panic("encode error: " + err.Error())
	}
	hash := makeHashNode(bz)

	fmt.Printf("FullNode: %s\n", n)
	fmt.Printf("Encoded: %X\n", bz)

	return hash
}

/** pulled in from ethereum trie/hasher.go **/

// keccak wraps sha3.state. In addition to the usual hash methods, it also supports
// Read to get a variable amount of data from the hash state. Read is faster than Sum
// because it doesn't copy the internal state, but also modifies the internal state.
type keccak interface {
	hash.Hash
	Read([]byte) (int, error)
}

func makeHashNode(data []byte) hashNode {
	h := sha3.NewLegacyKeccak256().(keccak)
	n := make(hashNode, h.Size())
	h.Write(data)
	h.Read(n)
	return n
}
