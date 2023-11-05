package services

import (
	"errors"
	"fmt"
	"github.com/alphabatem/btc_rune"
	"github.com/btcsuite/btcd/wire"
	"github.com/cloakd/common/context"
	"github.com/cloakd/common/services"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"os"
	"strconv"
	"time"
)

type HttpService struct {
	services.DefaultService

	BaseURL string
	Port    int

	startTime time.Time

	runeSvc *RuneService
	btcSvc  *BTCService
}

var ErrUnauthorized = errors.New("unauthorized")
var DeleteResponseOK = `{"status": 200, "error": ""}`

const HTTP_SVC = "http_svc"

func (svc HttpService) Id() string {
	return HTTP_SVC
}

func (svc *HttpService) Configure(ctx *context.Context) (err error) {
	svc.startTime = time.Now()
	svc.Port, err = strconv.Atoi(os.Getenv("HTTP_PORT"))
	if err != nil {
		return err
	}

	return svc.DefaultService.Configure(ctx)
}

func (svc *HttpService) Start() error {
	svc.runeSvc = svc.Service(RUNE_SVC).(*RuneService)
	svc.btcSvc = svc.Service(BTC_SVC).(*BTCService)
	r := gin.Default()

	r.Use(gin.Recovery())

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowCredentials = true
	config.AddAllowHeaders("Authorization")
	r.Use(cors.New(config))

	//Validation endpoints
	r.GET("/ping", svc.ping)

	btcG := r.Group("/btc")
	btcG.GET("/blocks", svc.btcBlocks)

	runeG := r.Group("/rune")
	runeG.GET("/mempool", svc.runeMempool)
	runeG.GET("/blocks/:id", svc.runeBlock)
	runeG.GET("/tx/:id", svc.runeTransaction)
	runeG.GET("/address/:id", svc.runeBalance)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})

	return r.Run(fmt.Sprintf(":%v", svc.Port))
}

type Pong struct {
	Message string `json:"message"`
}

// @Summary Ping service
// @Accept  json
// @Produce json
// @Success 200 {object} Pong "Returns pong if the service is up and running"
// @Router /ping [get]
func (svc *HttpService) ping(c *gin.Context) {
	c.JSON(200, Pong{"pong"})
}

func (svc *HttpService) btcBlocks(c *gin.Context) {
	resp, err := svc.btcSvc.RecentBlocks()
	if err != nil {
		c.AbortWithStatusJSON(400, err)
		return
	}

	hashes := make([]string, len(resp))
	for i, r := range resp {
		hashes[i] = r.String()
	}

	c.JSON(200, hashes)
}

type BlockHeader struct {
	Version          int32     `json:"version"`
	PrevBlock        string    `json:"prevBlock"`
	Timestamp        time.Time `json:"timestamp"`
	Bits             uint32    `json:"bits"`
	Nonce            uint32    `json:"nonce"`
	TransactionCount int       `json:"transactionCount"`
}

func (b *BlockHeader) FromWire(header *wire.BlockHeader) *BlockHeader {
	b.Version = header.Version
	b.PrevBlock = header.PrevBlock.String()
	b.Timestamp = header.Timestamp
	b.Bits = header.Bits
	b.Nonce = header.Nonce
	return b
}

type Block struct {
	Block            *BlockHeader            `json:"block"`
	RuneTransactions []*btc_rune.Transaction `json:"transactions"`
}

func (svc *HttpService) runeBlock(c *gin.Context) {
	block, transactions, err := svc.runeSvc.BlockTransactions(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(400, err)
		return
	}

	bh := BlockHeader{
		TransactionCount: len(block.Transactions),
	}
	bh.FromWire(&block.Header)

	c.JSON(200, &Block{
		Block:            &bh,
		RuneTransactions: transactions,
	})
}

type Txn struct {
	Transaction *wire.MsgTx           `json:"transaction"`
	Transfers   *btc_rune.Transaction `json:"transfers"`
}

func (svc *HttpService) runeTransaction(c *gin.Context) {
	tx, resp, err := svc.runeSvc.Transaction(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(400, err)
		return
	}
	c.JSON(200, &Txn{
		Transaction: tx,
		Transfers:   resp,
	})
}

func (svc *HttpService) runeBalance(c *gin.Context) {
	resp, err := svc.runeSvc.Balance(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(400, err)
		return
	}
	c.JSON(200, resp)
}

func (svc *HttpService) runeMempool(c *gin.Context) {
	c.JSON(200, Pong{"pong"})
}
