package services

import (
	"bytes"
	"encoding/binary"
	"github.com/alphabatem/btc_rune"
	"github.com/babilu-online/common/context"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcwallet/wallet"
	"io"
	"os"
)

type BTCService struct {
	context.DefaultService

	httpClient *rpcclient.Client

	latestBlocks []*chainhash.Hash
}

const BTC_SVC = "btc_svc"

func (svc BTCService) Id() string {
	return BTC_SVC
}

func (svc *BTCService) Start() (err error) {
	svc.latestBlocks = []*chainhash.Hash{}

	svc.httpClient, err = rpcclient.New(&rpcclient.ConnConfig{
		User: os.Getenv("RPC_USER"),
		Pass: os.Getenv("RPC_PASS"),
		Host: os.Getenv("RPC_URL"),
		//DisableTLS:   true,
		HTTPPostMode: true,
	}, nil)
	if err != nil {
		return err
	}

	return nil
}

func (svc *BTCService) RecentBlocks() ([]*chainhash.Hash, error) {
	blockCount, err := svc.httpClient.GetBlockCount()
	if err != nil {
		return nil, err
	}

	limit := 10
	if len(svc.latestBlocks) > limit {
		svc.latestBlocks = svc.latestBlocks[:limit-1]
	}

	if len(svc.latestBlocks) > 1 {
		hash, err := svc.httpClient.GetBestBlockHash()
		if err != nil {
			return nil, err
		}

		if !svc.latestBlocks[len(svc.latestBlocks)-1].IsEqual(hash) {
			svc.latestBlocks = append(svc.latestBlocks, hash)
		}

		return svc.latestBlocks, nil
	}

	//Seed from scratch
	start := blockCount - int64(limit)
	for i := start; i <= blockCount; i++ {
		hash, err := svc.httpClient.GetBlockHash(i)
		if err != nil {
			continue
		}
		svc.latestBlocks = append(svc.latestBlocks, hash)
	}

	return svc.latestBlocks, nil
}

func (svc *BTCService) Block(blockHash *chainhash.Hash) (*wire.MsgBlock, error) {
	return svc.httpClient.GetBlock(blockHash)
}

func (svc *BTCService) Transaction(txHash *chainhash.Hash) (*btcutil.Tx, error) {
	return svc.httpClient.GetRawTransaction(txHash)
}

func (svc *BTCService) CreateIssuanceTransaction(symbol string, decimals uint64) (*wire.MsgTx, error) {
	// Create a new transaction
	tx := wire.NewMsgTx(wire.TxVersion)

	// Create the TxIn (what you are spending)
	outpoint := wire.NewOutPoint(&chainhash.Hash{}, 1)
	txin := wire.NewTxIn(outpoint, nil, nil)
	tx.AddTxIn(txin)

	// Prepare Rune protocol issuance data
	var runeData bytes.Buffer
	runeData.WriteByte('R')

	// Encode SYMBOL and DECIMALS as Varints
	symbolEncoded := []byte(symbol)
	for _, s := range symbolEncoded {
		runeData.Write(svc.toVarInt(uint64(s - 'A')))
	}
	runeData.Write(svc.toVarInt(decimals))

	// Build OP_RETURN script
	opReturnScript, err := txscript.NewScriptBuilder().AddOp(txscript.OP_RETURN).AddData(runeData.Bytes()).Script()
	if err != nil {
		return nil, err
	}

	// Create the OP_RETURN TxOut
	opReturnTxOut := wire.NewTxOut(0, opReturnScript)
	tx.AddTxOut(opReturnTxOut)

	// Create the TxOut (where you are sending to)
	addr, _ := btcutil.DecodeAddress("recipientAddressHere", &chaincfg.MainNetParams)
	script, _ := txscript.PayToAddrScript(addr)
	txout := wire.NewTxOut(50000, script)
	tx.AddTxOut(txout)

	return tx, nil
}

func (svc *BTCService) CreateTransferTransaction(assignments []*btc_rune.Assignment) (*wire.MsgTx, error) {
	// Create a new transaction
	tx := wire.NewMsgTx(wire.TxVersion)

	// Create the TxIn (what you are spending)
	outpoint := wire.NewOutPoint(&chainhash.Hash{}, 1)
	txin := wire.NewTxIn(outpoint, nil, nil)
	tx.AddTxIn(txin)

	// Prepare Rune protocol transfer data
	var runeData bytes.Buffer
	runeData.WriteByte('R')

	// Encode assignments as Varints
	for _, a := range assignments {
		runeData.Write(svc.toVarInt(a.ID))
		runeData.Write(svc.toVarInt(a.Output))
		runeData.Write(svc.toVarInt(a.Amount))
	}

	// Build OP_RETURN script
	opReturnScript, err := txscript.NewScriptBuilder().AddOp(txscript.OP_RETURN).AddData(runeData.Bytes()).Script()
	if err != nil {
		return nil, err
	}

	// Create the OP_RETURN TxOut
	opReturnTxOut := wire.NewTxOut(0, opReturnScript)
	tx.AddTxOut(opReturnTxOut)

	// Create the TxOut (where you are sending to)
	addr, _ := btcutil.DecodeAddress("recipientAddressHere", &chaincfg.MainNetParams)
	script, _ := txscript.PayToAddrScript(addr)
	txout := wire.NewTxOut(50000, script)
	tx.AddTxOut(txout)

	return tx, nil
}

// Sign Signs the message with the given signer
func (svc *BTCService) Sign(tx *wire.MsgTx, signer []byte) ([]byte, error) {
	// Sign the transaction
	privateKey, _ := btcec.PrivKeyFromBytes(signer)
	sourceTx := &wire.MsgTx{} // This should be the transaction that your input is spending from
	sourceTxIndex := 0

	sigScript, _ := txscript.SignatureScript(sourceTx, sourceTxIndex, tx.TxOut[0].PkScript, txscript.SigHashAll, privateKey, true)
	tx.TxIn[0].SignatureScript = sigScript

	// Serialize the transaction
	var buf bytes.Buffer
	err := tx.Serialize(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func Send(tx *wire.MsgTx) error {
	loader := wallet.NewLoader(&chaincfg.MainNetParams, "./", true, wallet.DefaultDBTimeout, 0)
	w, err := loader.OpenExistingWallet([]byte("your-wallet-password"), false)
	if err != nil {
		return err
	}

	// Publish the transaction
	err = w.PublishTransaction(tx, "")
	if err != nil {
		return err
	}

	return nil
}

// Helper function to encode integers into Variants
func (svc *BTCService) toVarInt(value uint64) []byte {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, value)
	return buf[:n]
}

func (svc *BTCService) DecodeVarByte(buf *bytes.Buffer) []byte {
	prefix, err := buf.ReadByte()
	if err == io.EOF {
		return []byte{}
	}

	if prefix < 1 {
		next, err := buf.ReadByte()
		if err == io.EOF {
			return []byte{prefix}
		}

		return []byte{next}
	}

	//log.Println("PREFIX", prefix, buf.Len())

	if buf.Len() == 0 {
		return []byte{prefix}
	}

	if prefix == 254 {
		prefix = 6
	}

	if prefix == 255 {
		prefix = 8
	}

	by := make([]byte, prefix)
	_, _ = buf.Read(by)
	return by
}

func (svc *BTCService) DecodeVarInt(buf *bytes.Buffer) uint64 {
	by := svc.DecodeVarByte(buf)
	return svc.ByteToInt(by)
}

func (svc *BTCService) ByteToInt(by []byte) uint64 {
	switch {
	case len(by) == 0:
		return 0
	case len(by) == 1:
		return uint64(by[0])
	case len(by) <= 2:
		return uint64(binary.LittleEndian.Uint16(by))
	case len(by) <= 4:
		return uint64(binary.LittleEndian.Uint32(by))
	default:
		return binary.LittleEndian.Uint64(by)
	}
}
