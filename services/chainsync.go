package services

import (
	"encoding/json"
	"github.com/cloakd/common/services"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btclog"
	"log"
	"os"
)

type ChainSyncService struct {
	services.DefaultService

	btc  *BTCService
	rune *RuneService

	wsClient   *rpcclient.Client
	httpClient *rpcclient.Client

	blockHashes chan *chainhash.Hash
}

const CHAIN_SYNC_SVC = "chain_sync_svc"

func (svc ChainSyncService) Id() string {
	return CHAIN_SYNC_SVC
}

func (svc *ChainSyncService) Start() (err error) {
	svc.btc = svc.Service(BTC_SVC).(*BTCService)
	svc.rune = svc.Service(RUNE_SVC).(*RuneService)

	svc.blockHashes = make(chan *chainhash.Hash, 10)

	logger := btclog.NewBackend(os.Stdout).Logger("MAIN")
	logger.SetLevel(btclog.LevelDebug)
	rpcclient.UseLogger(logger)

	svc.httpClient, err = rpcclient.New(&rpcclient.ConnConfig{
		User:         "93c9cf43bb7263123fef469f13e695d23fca3c23",
		Pass:         "93c9cf43bb7263123fef469f13e695d23fca3c23",
		Host:         os.Getenv("RPC_URL"),
		DisableTLS:   true,
		HTTPPostMode: true,
	}, nil)
	if err != nil {
		return err
	}

	//th, _ := chainhash.NewHashFromStr("1aa98283f61cea9125aea58441067baca2533e2bbf8218b5e4f9ef7b8c0d8c30")
	//th, _ := chainhash.NewHashFromStr("2aefe2887654b3e4e7addd8f7c6496c26110833342830c19babda8d3875072ea")
	//tx, err := svc.httpClient.GetRawTransaction(th)
	//if err != nil {
	//	return err
	//}
	//
	//for _, out := range tx.MsgTx().TxOut {
	//	_ = svc.onRuneTransaction(*th, out.PkScript)
	//}

	//return svc.startWS()
	return nil
}

func (svc *ChainSyncService) startWS() (err error) {
	connCfg := &rpcclient.ConnConfig{
		User:         "93c9cf43bb7263123fef469f13e695d23fca3c23",
		Pass:         "93c9cf43bb7263123fef469f13e695d23fca3c23",
		Host:         os.Getenv("RPC_URL"),
		HTTPPostMode: false,
		DisableTLS:   true,
	}

	svc.wsClient, err = rpcclient.New(connCfg, &rpcclient.NotificationHandlers{
		OnClientConnected: func() {
			log.Println("Connected WSS")

			log.Println("Listening for Transactions")
			err := svc.wsClient.NotifyNewTransactions(false)
			if err != nil {
				panic(err)
			}

			log.Println("Listening for Blocks")
			err = svc.wsClient.NotifyBlocks()
			if err != nil {
				panic(err)
			}

		},
		OnFilteredBlockConnected: svc.onBlockConnected,
		OnRelevantTxAccepted:     svc.onRelevantTxAccepted,
		OnTxAccepted:             svc.onTxAccepted,
		OnUnknownNotification: func(method string, params []json.RawMessage) {
			log.Printf("Unknown Notification: %v", params)
		},
	})
	if err != nil {
		return err
	}

	go svc.listen() //TODO Remove?
	return nil
}

func (svc *ChainSyncService) onBlockConnected(height int32, header *wire.BlockHeader, txs []*btcutil.Tx) {
	log.Printf("Block Connected: %v - %s - Len: %v", height, header.BlockHash(), len(txs))
}

func (svc *ChainSyncService) onTxAccepted(hash *chainhash.Hash, amount btcutil.Amount) {
	log.Printf("New TXN: %s", hash)
	err := svc.handleNewBlock(hash)
	if err != nil {
		log.Printf("onTxAccepted Err: %s", err)
	}
}

func (svc *ChainSyncService) onRelevantTxAccepted(transaction []byte) {
	log.Printf("Txn: %v", transaction)
}

func (svc *ChainSyncService) listen() {
	for {
		select {
		case block := <-svc.blockHashes:
			err := svc.handleNewBlock(block)
			if err != nil {
				log.Println("handleBlockErr", err)
			}

		}
	}
}

func (svc *ChainSyncService) handleNewBlock(blockHash *chainhash.Hash) error {
	log.Println("New Block", blockHash)
	block, err := svc.httpClient.GetBlock(blockHash)
	if err != nil {
		return err
	}

	for _, tx := range block.Transactions {
		isRuneTx, runeData := svc.rune.isRuneTransaction(tx)
		if !isRuneTx {
			continue
		}

		err = svc.onRuneTransaction(tx.TxHash(), runeData)
		if err != nil {
			return err
		}
	}
	return nil
}

func (svc *ChainSyncService) onRuneTransaction(txID chainhash.Hash, data []byte) error {
	tx := svc.rune.DecodeTransaction(&txID, data)
	if tx == nil {
		return nil
	}

	if tx.Issuance != nil {
		log.Printf("ISSUANCE: %+v\n", tx.Issuance)
	}

	for _, a := range tx.Transfers {
		log.Printf("XFER: %v - %v - %v", a.ID, a.Amount, a.Output)
	}

	return nil
}
