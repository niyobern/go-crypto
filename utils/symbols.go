package utils

type TrieNode struct {
	children map[rune]*TrieNode
	isEnd    bool
}

func newTrieNode() *TrieNode {
	return &TrieNode{children: make(map[rune]*TrieNode)}
}

func (t *TrieNode) insert(word string) {
	current := t
	for _, ch := range word {
		if _, found := current.children[ch]; !found {
			current.children[ch] = newTrieNode()
		}
		current = current.children[ch]
	}
	current.isEnd = true
}

func (t *TrieNode) searchPrefix(prefix string) string {
	current := t
	var lastMatch string

	for _, ch := range prefix {
		if node, found := current.children[ch]; found {
			current = node
			lastMatch += string(ch)
			if current.isEnd {
				return lastMatch
			}
		} else {
			break
		}
	}
	return ""
}

func GetQuote(symbol string, trie *TrieNode) string {
	prefix := trie.searchPrefix(symbol)
	if prefix == "" || symbol[len(prefix):] != "USDT" {
		return ""
	}
	return prefix + "-" + symbol[len(prefix):]
}

func Initialize() *TrieNode {
	trie := newTrieNode()
	quotes := []string{"XEC", "KAVA", "EDU", "AR", "HNT", "LUNA", "BSW", "BNC", "FLOW", "OP", "LISTA", "NTRN", "NVT", "HARD", "QI", "ZK", "VOXEL", "DYM", "TWT", "MBL", "ATEM", "BURGER", "TFUEL", "XAI", "NFP", "MAGIC", "CKB", "MOVR", "APT", "SUI", "RENDER", "NOT", "GMX", "GNS", "ROSE", "TIA", "EPX", "IO", "USTC", "OSMO", "TAO", "TNSR", "BONK", "SEI", "IOTA", "WIF", "JST", "ETHW", "FLR", "NEAR", "MANTA", "RUNE", "CELO", "SXP", "PYTH", "HBAR", "KLAY", "JTO", "KDA", "EGLD", "GFT", "STRAX", "VELO", "XYM", "NFT", "BOME", "JUP", "SCRT", "POLYX", "XNO", "ALPINE", "BB"}

	for _, quote := range quotes {
		trie.insert(quote)
	}
	return trie
}