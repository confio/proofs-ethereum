package proof

import (
	"fmt"

	"github.com/ethereum/go-ethereum/rlp"
)

var hshr = newHasher(0, 0, nil)

func collapseShortNode(n *shortNode) *shortNode {
	collapsed := n.copy()
	collapsed.Key = hexToCompact(n.Key)
	return collapsed
}

func hashShortNode(n *shortNode) []byte {
	if _, ok := n.Val.(valueNode); !ok {
		panic("only implemented for value nodes")
	}

	bz, err := rlp.EncodeToBytes(collapseShortNode(n))
	if err != nil {
		panic("encode error: " + err.Error())
	}
	hash := hshr.makeHashNode(bz)

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

func hashFullNode(n *fullNode) []byte {
	// this is pre-processing
	collapsed := n.copy()
	for i := 0; i < 16; i++ {
		switch child := collapsed.Children[i].(type) {
		case *shortNode:
			collapsed.Children[i] = collapseShortNode(child)
		case *fullNode:
			collapsed.Children[i] = hashNode(hashFullNode(child))
			// leave valueNode and hashNode (or reference) untouched
		}
	}

	// this is encoding process
	bz, err := rlp.EncodeToBytes(collapsed)
	if err != nil {
		panic("encode error: " + err.Error())
	}
	hash := hshr.makeHashNode(bz)

	fmt.Printf("FullNode: %s\n", n)
	fmt.Printf("Encoded: %X\n", bz)

	return hash
}

// try to copy the ethereum code to get anything working
func ethHashFullNode(n *fullNode) []byte {
	h, _, err := hshr.hash(n, nil, false)
	if err != nil {
		panic(err)
	}
	return h.(hashNode)
}
