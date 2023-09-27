package btc_rune

type UTXO struct {
	Owner  string `json:"owner"`
	Amount uint64 `json:"amount"`
	RuneID uint64 `json:"runeID"`
}
