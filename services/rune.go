package services

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/alphabatem/btc_rune"
	"github.com/babilu-online/common/context"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"log"
	"math/big"
	"strconv"
	"strings"
)

type RuneService struct {
	context.DefaultService

	btc *BTCService
}

const RUNE_SVC = "rune_svc"

func (svc RuneService) Id() string {
	return RUNE_SVC
}

func (svc *RuneService) Start() error {
	svc.btc = svc.Service(BTC_SVC).(*BTCService)

	return nil
}

//TODO Complete

func (svc *RuneService) Issue(signer []byte, symbol string, decimals, amount uint64) error {
	tx, err := svc.btc.CreateIssuanceTransaction(symbol, decimals)
	if err != nil {
		return err
	}

	_, err = svc.btc.Sign(tx, signer)
	if err != nil {
		return err
	}

	log.Println(tx)

	return nil
}

func (svc *RuneService) Transfer(signer []byte, rune *btc_rune.Rune, assignments []*btc_rune.Assignment) error {

	tx, err := svc.btc.CreateTransferTransaction(assignments)
	if err != nil {
		return err
	}

	_, err = svc.btc.Sign(tx, signer)
	if err != nil {
		return err
	}

	log.Println(tx)
	return nil
}

func (svc *RuneService) Balance(addr string) (map[string]int64, error) {
	balances := map[string]int64{}

	txns, err := svc.btc.httpClient.ListTransactions(addr)
	if err != nil {
		return nil, err
	}

	for _, tx := range txns {
		log.Println(tx.TimeReceived)
		//h, _ := chainhash.NewHashFromStr(tx.BlockHash)
		//ttx, err := svc.btc.Transaction(h)
		//if err != nil {
		//	continue
		//}

		//ttx.MsgTx().BtcEncode()
		//
		//if ok, data := svc.isRuneTransaction(ttx.MsgTx()); ok {
		//	rtx := svc.DecodeTransaction(h, data)
		//
		//}

	}

	return balances, nil
}

func (svc *RuneService) Transaction(txHash string) (*wire.MsgTx, *btc_rune.Transaction, error) {
	hash, err := chainhash.NewHashFromStr(txHash)
	if err != nil {
		log.Println("Invalid hash")
		return nil, nil, err
	}

	tx, err := svc.btc.Transaction(hash)
	if err != nil {
		log.Println("TXN Not Found")
		return nil, nil, err
	}

	ok, data := svc.isRuneTransaction(tx.MsgTx())
	if !ok {
		return nil, nil, errors.New("not a rune tx")
	}

	return tx.MsgTx(), svc.DecodeTransaction(hash, data), nil
}

func (svc *RuneService) BlockTransactions(blockHash string) (*wire.MsgBlock, []*btc_rune.Transaction, error) {
	hash, err := chainhash.NewHashFromStr(blockHash)
	if err != nil {
		log.Println("Invalid hash")
		return nil, nil, err
	}

	block, err := svc.btc.Block(hash)
	if err != nil {
		return nil, nil, err
	}

	var txns []*btc_rune.Transaction
	for _, t := range block.Transactions {
		ok, data := svc.isRuneTransaction(t)
		if !ok {
			continue
		}

		txHash := t.TxHash()
		txns = append(txns, svc.DecodeTransaction(&txHash, data))
	}

	return block, txns, nil
}

func (svc *RuneService) isRuneTransaction(tx *wire.MsgTx) (bool, []byte) {
	for _, txOut := range tx.TxOut {
		pkScript := txOut.PkScript
		if txscript.IsPayToWitnessPubKeyHash(pkScript) || txscript.IsPayToPubKeyHash(pkScript) {
			continue
		}

		// Check for OP_RETURN and ASCII 'R'
		if len(pkScript) > 2 && pkScript[0] == txscript.OP_RETURN && pkScript[1] == 1 && pkScript[2] == 'R' {
			return true, pkScript
		}
	}
	return false, nil
}

func (svc *RuneService) DecodeTransaction(txHash *chainhash.Hash, runeData []byte) *btc_rune.Transaction {
	buffer := bytes.NewBuffer(runeData)

	cmd, _ := binary.ReadUvarint(buffer)
	dataType := svc.btc.DecodeVarInt(buffer)

	if cmd != txscript.OP_RETURN && dataType != 82 {
		return nil //Not an OP_RETURN
	}

	tx := btc_rune.Transaction{
		Hash: txHash.String(),
	}

	i := 0
	for buffer.Len() > 0 {
		size, _ := binary.ReadUvarint(buffer)
		msg := make([]byte, size)
		_, _ = buffer.Read(msg)

		if i == 0 { //Transfer
			tx.Transfers = svc.decodeTransfer(msg)
		} else { //Issuance
			tx.Issuance = svc.decodeIssuance(msg)
		}

		i++
	}
	return &tx
}

func (svc *RuneService) decodeIssuance(data []byte) *btc_rune.Rune {
	log.Println("Issuance", data)
	buffer := bytes.NewBuffer(data)

	symbol := svc.btc.DecodeVarByte(buffer)
	//log.Println("BYT", symbol)

	decimals := svc.btc.DecodeVarInt(buffer)
	//log.Println("decimals", decimals)

	return &btc_rune.Rune{
		Symbol:   svc.IntToBase26(symbol),
		Decimals: decimals,
	}
}

func (svc *RuneService) decodeTransfer(data []byte) []*btc_rune.Assignment {
	buffer := bytes.NewBuffer(data)
	var assignments []*btc_rune.Assignment

	for buffer.Len() >= 9 {
		assignments = append(assignments, &btc_rune.Assignment{
			ID:     svc.btc.DecodeVarInt(buffer),
			Output: svc.btc.DecodeVarInt(buffer),
			Amount: svc.btc.DecodeVarInt(buffer),
		})
	}

	return assignments
}

func (svc *RuneService) hexToBase26(hexStr string) string {
	hexNum := new(big.Int)
	hexNum.SetString(hexStr, 16)
	var base26StrBuilder strings.Builder

	base26Chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	for hexNum.Cmp(big.NewInt(0)) > 0 {
		remainder := new(big.Int)
		remainder.Mod(hexNum, big.NewInt(26))
		hexNum.Div(hexNum, big.NewInt(26))
		base26StrBuilder.WriteByte(base26Chars[remainder.Int64()])
	}

	return base26StrBuilder.String()
}

func (svc *RuneService) IntToBase26(txt []byte) string {
	base26Chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	if len(txt) < 4 {
		txt = append(txt, 0, 0, 0, 0)
	}

	sym := fmt.Sprintf("%v", binary.LittleEndian.Uint32(txt))
	if len(sym) < 2 {
		return ""
	}

	if len(sym)%2 != 0 {
		log.Println("IntToBase26 - Invalid Char Amount", sym, txt)
		return ""
	}
	log.Println("IntToBase26", sym, txt)
	var text strings.Builder
	for i := 0; i < len(sym); i += 2 {
		n, _ := strconv.Atoi(sym[i : i+2])
		if n > 26 {
			log.Println("Skipping", n)
			continue
		}
		text.WriteByte(base26Chars[n])
	}
	return text.String()
}
