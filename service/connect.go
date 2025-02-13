package service

import (
	"context"
	"net/url"

	"github.com/pkg/errors"

	"github.com/libsv/payd"
	"github.com/libsv/payd/config"
)

type connect struct {
	wtr    payd.ConnectWriter
	invRdr payd.InvoiceReader
	dppCfg *config.DPP
}

// NewConnect will setup a new connect service used to connect this wallet to a dpp socket server.
func NewConnect(wtr payd.ConnectWriter, invRdr payd.InvoiceReader, dppCfg *config.DPP) *connect {
	return &connect{
		wtr:    wtr,
		invRdr: invRdr,
		dppCfg: dppCfg,
	}
}

// Connect will connect this server to a third party service using either a socket or peer channel protocol.
func (c *connect) Connect(ctx context.Context, args payd.ConnectArgs) error {
	if err := args.Validate(); err != nil {
		return err
	}
	// get the invoice if an error then it isn't here.
	if _, err := c.invRdr.Invoice(ctx, payd.InvoiceArgs{InvoiceID: args.InvoiceID}); err != nil {
		return errors.Wrapf(err, "failed to validate invoice %s when attempting to create connection", args.InvoiceID)
	}
	u, err := url.Parse(c.dppCfg.ServerHost)
	if err != nil {
		return errors.Wrap(err, "failed to parse url")
	}
	switch u.Scheme {
	case "ws", "wss":
		return c.wtr.Connect(ctx, args)
	}
	return nil
}
