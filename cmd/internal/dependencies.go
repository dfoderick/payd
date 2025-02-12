package internal

import (
	"context"
	"net/http"
	"time"

	"github.com/libsv/payd/log"

	"github.com/jmoiron/sqlx"
	"github.com/libsv/go-bc/spv"
	"github.com/theflyingcodr/sockets/client"
	"github.com/tonicpow/go-minercraft"

	"github.com/libsv/payd"
	"github.com/libsv/payd/config"
	dataHttp "github.com/libsv/payd/data/http"
	"github.com/libsv/payd/data/mapi"
	dsoc "github.com/libsv/payd/data/sockets"
	paydSQL "github.com/libsv/payd/data/sqlite"
	"github.com/libsv/payd/service"
)

// RestDeps contains all dependencies used for the rest client.
type RestDeps struct {
	DestinationService    payd.DestinationsService
	PaymentService        payd.PaymentsService
	PaymentRequestService payd.PaymentRequestService
	PayService            payd.PayService
	EnvelopeService       payd.EnvelopeService
	InvoiceService        payd.InvoiceService
	BalanceService        payd.BalanceService
	ProofService          payd.ProofsService
	OwnerService          payd.OwnerService
	UserService           payd.UserService
	TransactionService    payd.TransactionService
}

// SetupRestDeps will setup dependencies used in the rest server.
func SetupRestDeps(cfg *config.Config, l log.Logger, db *sqlx.DB, c *client.Client) *RestDeps {
	mapiCli, err := minercraft.NewClient(nil, nil, []*minercraft.Miner{
		{
			Name:  cfg.Mapi.MinerName,
			Token: cfg.Mapi.Token,
			URL:   cfg.Mapi.URL,
		},
	})
	if err != nil {
		l.Fatal(err, "failed to setup mapi client")
	}
	sqlLiteStore := paydSQL.NewSQLiteStore(db)
	proofSvc := service.NewProofsService(sqlLiteStore)

	pcSvc := service.NewPeerChannelsSvc(sqlLiteStore, cfg.PeerChannels)
	pcNotifSvc := service.NewPeerChannelsNotifyService(cfg.PeerChannels, pcSvc)
	pcNotifSvc.RegisterHandler(payd.PeerChannelHandlerTypeProof, proofSvc)

	mapiStore := mapi.NewMapi(cfg.Mapi, mapiCli)
	spvv, err := spv.NewPaymentVerifier(dataHttp.NewHeaderSVConnection(&http.Client{Timeout: time.Duration(cfg.HeadersClient.Timeout) * time.Second}, cfg.HeadersClient.Address))
	if err != nil {
		l.Fatal(err, "failed to create spv client")
	}

	spvc, err := spv.NewEnvelopeCreator(sqlLiteStore, sqlLiteStore)
	if err != nil {
		l.Fatal(err, "failed to create spv verifier")
	}

	seedSvc := service.NewSeedService()
	privKeySvc := service.NewPrivateKeys(sqlLiteStore, cfg.Wallet.Network == "mainnet")
	destSvc := service.NewDestinationsService(cfg.Wallet, privKeySvc, sqlLiteStore, sqlLiteStore, sqlLiteStore, seedSvc)
	paymentSvc := service.NewPayments(l, spvv, sqlLiteStore, sqlLiteStore, sqlLiteStore, &paydSQL.Transacter{}, mapiStore, sqlLiteStore, sqlLiteStore, pcSvc, pcNotifSvc, cfg.PeerChannels)
	envSvc := service.NewEnvelopes(privKeySvc, sqlLiteStore, sqlLiteStore, sqlLiteStore, seedSvc, spvc)
	paySvc := service.NewPayStrategy().Register(
		service.NewPayService(&paydSQL.Transacter{}, dataHttp.NewDPP(&http.Client{Timeout: time.Duration(cfg.DPP.Timeout) * time.Second}), envSvc, cfg.Server, pcNotifSvc, sqlLiteStore, sqlLiteStore, cfg.Wallet),
		"http", "https",
	).Register(
		service.NewPayChannel(dsoc.NewPaymentChannel(*cfg.Socket, c)), "ws", "wss",
	)
	paymentReqSvc := service.NewPaymentRequest(cfg.Wallet, destSvc, mapiStore, sqlLiteStore, sqlLiteStore)
	invoiceSvc := service.NewInvoice(cfg.Server, cfg.Wallet, sqlLiteStore, destSvc, &paydSQL.Transacter{}, service.NewTimestampService())
	balanceSvc := service.NewBalance(sqlLiteStore)
	connectService := service.NewConnect(dsoc.NewConnect(cfg.DPP, c), invoiceSvc, cfg.DPP)
	invoiceSvc.SetConnectionService(connectService)
	ownerSvc := service.NewOwnerService(sqlLiteStore)
	userSvc := service.NewUsersService(sqlLiteStore, privKeySvc)

	transactionService := service.NewTransactions(&paydSQL.Transacter{}, sqlLiteStore, sqlLiteStore, sqlLiteStore)

	// create master private key if it doesn't exist
	if err = privKeySvc.Create(context.Background(), "masterkey", 1); err != nil {
		l.Fatal(err, "failed to create master key")
	}

	return &RestDeps{
		DestinationService:    destSvc,
		PaymentService:        paymentSvc,
		PaymentRequestService: paymentReqSvc,
		PayService:            paySvc,
		EnvelopeService:       envSvc,
		InvoiceService:        invoiceSvc,
		BalanceService:        balanceSvc,
		ProofService:          proofSvc,
		OwnerService:          ownerSvc,
		UserService:           userSvc,
		TransactionService:    transactionService,
	}
}

// SocketDeps contains all dependencies of the socket server.
type SocketDeps struct {
	DestinationService        payd.DestinationsService
	PaymentService            payd.PaymentsService
	PayService                payd.PayService
	EnvelopeService           payd.EnvelopeService
	InvoiceService            payd.InvoiceService
	BalanceService            payd.BalanceService
	ProofService              payd.ProofsService
	OwnerService              payd.OwnerService
	PaymentRequestService     payd.PaymentRequestService
	ConnectService            payd.ConnectService
	TransactionService        payd.TransactionService
	PeerChannelsService       payd.PeerChannelsService
	PeerChannelsNotifyService payd.PeerChannelsNotifyService
}

// SetupSocketDeps will setup dependencies used in the socket server.
func SetupSocketDeps(cfg *config.Config, l log.Logger, db *sqlx.DB, c *client.Client) *SocketDeps {
	mapiCli, err := minercraft.NewClient(nil, nil, []*minercraft.Miner{
		{
			Name:  cfg.Mapi.MinerName,
			Token: cfg.Mapi.Token,
			URL:   cfg.Mapi.URL,
		},
	})
	if err != nil {
		l.Fatal(err, "failed to setup mapi client")
	}
	sqlLiteStore := paydSQL.NewSQLiteStore(db)
	proofSvc := service.NewProofsService(sqlLiteStore)
	pcSvc := service.NewPeerChannelsSvc(sqlLiteStore, cfg.PeerChannels)
	pcNotifSvc := service.NewPeerChannelsNotifyService(cfg.PeerChannels, pcSvc)
	pcNotifSvc.RegisterHandler(payd.PeerChannelHandlerTypeProof, proofSvc)
	mapiStore := mapi.NewMapi(cfg.Mapi, mapiCli)
	spvv, err := spv.NewPaymentVerifier(dataHttp.NewHeaderSVConnection(&http.Client{Timeout: time.Duration(cfg.HeadersClient.Timeout) * time.Second}, cfg.HeadersClient.Address))
	if err != nil {
		l.Fatal(err, "failed to create spv client")
	}

	spvc, err := spv.NewEnvelopeCreator(sqlLiteStore, sqlLiteStore)
	if err != nil {
		l.Fatal(err, "failed to create spv verifier")
	}

	seedSvc := service.NewSeedService()
	privKeySvc := service.NewPrivateKeys(sqlLiteStore, cfg.Wallet.Network == "mainnet")
	destSvc := service.NewDestinationsService(cfg.Wallet, privKeySvc, sqlLiteStore, sqlLiteStore, sqlLiteStore, seedSvc)
	paymentSvc := service.NewPayments(l, spvv, sqlLiteStore, sqlLiteStore, sqlLiteStore, &paydSQL.Transacter{}, mapiStore, sqlLiteStore, sqlLiteStore, pcSvc, pcNotifSvc, cfg.PeerChannels)
	envSvc := service.NewEnvelopes(privKeySvc, sqlLiteStore, sqlLiteStore, sqlLiteStore, seedSvc, spvc)
	paySvc := service.NewPayStrategy().Register(
		service.NewPayService(&paydSQL.Transacter{}, dataHttp.NewDPP(&http.Client{Timeout: time.Duration(cfg.DPP.Timeout) * time.Second}), envSvc, cfg.Server, pcNotifSvc, sqlLiteStore, sqlLiteStore, cfg.Wallet),
		"http", "https",
	).Register(service.NewPayChannel(dsoc.NewPaymentChannel(*cfg.Socket, c)), "ws", "wss")
	invoiceSvc := service.NewInvoice(cfg.Server, cfg.Wallet, sqlLiteStore, destSvc, &paydSQL.Transacter{}, service.NewTimestampService())
	balanceSvc := service.NewBalance(sqlLiteStore)
	ownerSvc := service.NewOwnerService(sqlLiteStore)
	paymentReqSvc := service.NewPaymentRequest(cfg.Wallet, destSvc, mapiStore, sqlLiteStore, sqlLiteStore)
	connectService := service.NewConnect(dsoc.NewConnect(cfg.DPP, c), invoiceSvc, cfg.DPP)
	invoiceSvc.SetConnectionService(connectService)
	transactionService := service.NewTransactions(&paydSQL.Transacter{}, sqlLiteStore, sqlLiteStore, sqlLiteStore)

	// create master private key if it doesn't exist
	if err = privKeySvc.Create(context.Background(), "masterkey", 1); err != nil {
		l.Fatal(err, "failed to create master key")
	}

	return &SocketDeps{
		DestinationService:        destSvc,
		PaymentService:            paymentSvc,
		PayService:                paySvc,
		EnvelopeService:           envSvc,
		InvoiceService:            invoiceSvc,
		BalanceService:            balanceSvc,
		ProofService:              proofSvc,
		OwnerService:              ownerSvc,
		PaymentRequestService:     paymentReqSvc,
		ConnectService:            connectService,
		TransactionService:        transactionService,
		PeerChannelsService:       pcSvc,
		PeerChannelsNotifyService: pcNotifSvc,
	}
}
