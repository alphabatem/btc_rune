package btc_rune

type Rune struct {
	Symbol   string `json:"symbol"`
	Decimals uint64 `json:"decimals"`
}

type Transaction struct {
	Hash      string    `json:"hash"`
	Issuance  *Rune     `json:"issuance,omitempty"`
	Transfers Transfers `json:"transfers"`
}

type Transfers []*Assignment
