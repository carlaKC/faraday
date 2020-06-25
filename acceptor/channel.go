package acceptor

import (
	"bytes"
	"context"
	"fmt"

	"sync"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/routing/route"
	"github.com/prometheus/common/log"
)

type acceptFunc func(*lnrpc.ChannelAcceptRequest) *lnrpc.ChannelAcceptResponse

type ChannelAcceptor struct {
	Accept      func(ctx context.Context, accept func(*lnrpc.ChannelAcceptRequest) *lnrpc.ChannelAcceptResponse) (chan error, error)
	AcceptPeers []route.Vertex
	RejectPeers []route.Vertex

	wg   sync.WaitGroup
	quit chan struct{}
}

func (c *ChannelAcceptor) Start() error {
	return c.ListBasedAccept()
}

func (c *ChannelAcceptor) ListBasedAccept() error {
	c.wg.Add(1)

	var accept acceptFunc

	switch {
	case len(c.AcceptPeers) != 0:
		accept = c.acceptPeers

	case len(c.RejectPeers) != 0:
		accept = c.rejectPeers

	default:
		return fmt.Errorf("no acceptance strategies configured")
	}

	ctx, cancel := context.WithCancel(context.Background())
	acceptorErrs, err := c.Accept(ctx, accept)
	if err != nil {
		return err
	}

	go func() {
		// Defer cancelling our context, this will cancel our acceptor
		// stream. If we fail or are given the instruction to quit, we
		// will cancel our stream.
		defer cancel()
		defer c.wg.Done()

		for {
			select {
			// Exit if we get the instruction to shut down.
			case <-c.quit:
				return

				// Accept requests until we fail.
			case err := <-acceptorErrs:
				if err != nil {
					log.Warnf("acceptor error: %v", err)
				}
				return
			}
		}
	}()

	return nil
}

func (c *ChannelAcceptor) acceptPeers(req *lnrpc.ChannelAcceptRequest) *lnrpc.ChannelAcceptResponse {
	accept := false
	for _, peer := range c.AcceptPeers {
		if bytes.Equal(peer[:], req.NodePubkey) {
			accept = true
		}
	}

	log.Infof("received request from: %v, accept: %v", req.NodePubkey, accept)

	return &lnrpc.ChannelAcceptResponse{
		Accept:          accept,
		PendingChanId:   req.PendingChanId,
		RejectionReason: "I don't know you",
	}
}

func (c *ChannelAcceptor) rejectPeers(req *lnrpc.ChannelAcceptRequest) *lnrpc.ChannelAcceptResponse {
	accept := true
	for _, peer := range c.RejectPeers {
		if bytes.Equal(peer[:], req.NodePubkey) {
			accept = false
		}
	}

	log.Infof("received request from: %v, accept: %v", req.NodePubkey, accept)

	return &lnrpc.ChannelAcceptResponse{
		Accept:          accept,
		PendingChanId:   req.PendingChanId,
		RejectionReason: "I don't like you",
	}
}

func (c *ChannelAcceptor) Stop() error {
	close(c.quit)
	c.wg.Wait()

	return nil
}
