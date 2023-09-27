package services

import (
	"encoding/binary"
	"encoding/hex"
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

func TestRuneService_DecodeTransaction(t *testing.T) {

	btc := BTCService{}
	err := btc.Start()
	if err != nil {
		t.Fatal(err)
	}

	svc := RuneService{
		btc: &btc,
	}

	hash, _ := chainhash.NewHashFromStr("2aefe2887654b3e4e7addd8f7c6496c26110833342830c19babda8d3875072ea")
	tx, err := btc.Transaction(hash)

	_, data := svc.isRuneTransaction(tx.MsgTx())

	txn := svc.DecodeTransaction(hash, data)
	log.Println(txn)

	log.Printf("%+v\n", txn.Issuance)

	for _, xfer := range txn.Transfers {
		log.Printf("%+v\n", xfer)
	}
}

func TestRuneService_DecodeTransaction_OridnalsWallet(t *testing.T) {

	btc := BTCService{}
	err := btc.Start()
	if err != nil {
		t.Fatal(err)
	}

	svc := RuneService{
		btc: &btc,
	}

	hash, _ := chainhash.NewHashFromStr("1aa98283f61cea9125aea58441067baca2533e2bbf8218b5e4f9ef7b8c0d8c30")
	tx, err := btc.Transaction(hash)
	if err != nil {
		t.Fatal(err)
	}

	_, data := svc.isRuneTransaction(tx.MsgTx())

	txn := svc.DecodeTransaction(hash, data)
	log.Println(txn)

	if txn.Issuance.Symbol != "RUNE" {
		t.Fatalf("Incorrect symbol name")
	}

	for _, xfer := range txn.Transfers {
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

	//i, err := strconv.ParseInt(h, 16, 64)
	//if err != nil {
	//	t.Fatal(err)
	//}

	log.Printf("%v", binary.LittleEndian.Uint32([]byte(h)))

	svc := RuneService{}

	out := svc.hexToBase26(h)
	log.Println(out)
	out = svc.ToBase26([]byte(h))
	log.Println(out)
}
