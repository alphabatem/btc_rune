package services

import (
	"encoding/binary"
	"encoding/hex"
	"github.com/alphabatem/btc_rune"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/joho/godotenv"
	"log"
	"testing"
)

func init() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

}

func TestRuneService_DecodeTransactionAll(t *testing.T) {
	hashPepe, _ := chainhash.NewHashFromStr("804c299bad4457daeab28c5227d36c3920d92b98dc73e4f37fe1497956d91469")
	hashRune, _ := chainhash.NewHashFromStr("1aa98283f61cea9125aea58441067baca2533e2bbf8218b5e4f9ef7b8c0d8c30")
	hashRevoFusion, _ := chainhash.NewHashFromStr("2aefe2887654b3e4e7addd8f7c6496c26110833342830c19babda8d3875072ea")
	hashTrevor, _ := chainhash.NewHashFromStr("92aa890ba62ecd3a7dee1184111b044da9569aed4ee2e877b0e9f7fb81ee91f2")
	hashChad, _ := chainhash.NewHashFromStr("4a57ea9f4e29e20e4486467643b0557b95a1e7b7c63eaa7961da972cc2818dd4")

	hashes := map[string]*chainhash.Hash{
		"PEPE":   hashPepe,
		"RUNE":   hashRune,
		"J":      hashRevoFusion,
		"TREVOR": hashTrevor,
		"CHAD":   hashChad,
	}

	for expectedSym, h := range hashes {
		tx := checkTransaction(t, h)
		if tx.Issuance.Symbol != expectedSym {
			t.Logf("Expected: %s - Got: %s", expectedSym, tx.Issuance.Symbol)
		}

	}

}

// https://twitter.com/YesChads/status/1706861200816906714
func TestRuneService_DecodeTransactionTrevor(t *testing.T) {
	expectedSym := "TREVOR"
	expectedDec := uint64(0)
	hash, _ := chainhash.NewHashFromStr("92aa890ba62ecd3a7dee1184111b044da9569aed4ee2e877b0e9f7fb81ee91f2")
	tx := checkTransaction(t, hash)
	if tx.Issuance.Symbol != expectedSym {
		t.Fatalf("Expected: %s - Got: %s", expectedSym, tx.Issuance.Symbol)
	}
	if tx.Issuance.Decimals != expectedDec {
		t.Fatalf("Expected: %v - Got: %v", expectedDec, tx.Issuance.Decimals)
	}

	for _, xfer := range tx.Transfers {
		log.Printf("%+v\n", xfer)
	}
}

func TestRuneService_DecodeTransactionPepe(t *testing.T) {
	expectedSym := "PEPE"
	expectedDec := uint64(18)
	hash, _ := chainhash.NewHashFromStr("804c299bad4457daeab28c5227d36c3920d92b98dc73e4f37fe1497956d91469")
	tx := checkTransaction(t, hash)
	if tx.Issuance.Symbol != expectedSym {
		t.Fatalf("Expected: %s - Got: %s", expectedSym, tx.Issuance.Symbol)
	}
	if tx.Issuance.Decimals != expectedDec {
		t.Fatalf("Expected: %v - Got: %v", expectedDec, tx.Issuance.Decimals)
	}

	for _, xfer := range tx.Transfers {
		log.Printf("%+v\n", xfer)
	}
}

// https://twitter.com/revofusion/status/1706533725230792974
func TestRuneService_DecodeTransaction_RevoFusion(t *testing.T) {
	expectedSym := "RUNE"
	expectedDec := uint64(0)
	hash, _ := chainhash.NewHashFromStr("2aefe2887654b3e4e7addd8f7c6496c26110833342830c19babda8d3875072ea")
	tx := checkTransaction(t, hash)
	if tx.Issuance.Symbol != expectedSym {
		t.Fatalf("Expected: %s - Got: %s", expectedSym, tx.Issuance.Symbol)
	}
	if tx.Issuance.Decimals != expectedDec {
		t.Fatalf("Expected: %v - Got: %v", expectedDec, tx.Issuance.Decimals)
	}

	for _, xfer := range tx.Transfers {
		log.Printf("%+v\n", xfer)
	}
}

func TestRuneService_DecodeTransaction_OridnalsWallet(t *testing.T) {
	expectedSym := "RUNE"
	expectedDec := uint64(18)
	hash, _ := chainhash.NewHashFromStr("1aa98283f61cea9125aea58441067baca2533e2bbf8218b5e4f9ef7b8c0d8c30")
	tx := checkTransaction(t, hash)
	if tx.Issuance.Symbol != expectedSym {
		t.Fatalf("Expected: %s - Got: %s", expectedSym, tx.Issuance.Symbol)
	}
	if tx.Issuance.Decimals != expectedDec {
		t.Fatalf("Expected: %v - Got: %v", expectedDec, tx.Issuance.Decimals)
	}

	for _, xfer := range tx.Transfers {
		log.Printf("%+v\n", xfer)
	}
}

func TestRuneService_ToBase26(t *testing.T) {
	h := "b50c05"

	byt, err := hex.DecodeString(h)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Byt", byt)

	log.Printf("%v", binary.LittleEndian.Uint32([]byte(h)))

	svc := RuneService{}

	out := svc.hexToBase26(h)
	log.Println(out)
}

func checkTransaction(t *testing.T, hash *chainhash.Hash) *btc_rune.Transaction {
	btc := BTCService{}
	err := btc.Start()
	if err != nil {
		t.Fatal(err)
	}

	svc := RuneService{
		btc: &btc,
	}

	tx, err := btc.Transaction(hash)
	if err != nil {
		t.Fatal(err)
	}

	_, data := svc.isRuneTransaction(tx.MsgTx())

	return svc.DecodeTransaction(hash, data)
}
