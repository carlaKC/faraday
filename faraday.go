// Package faraday contains the main function for faraday.
package faraday

import (
	"fmt"

	"github.com/lightninglabs/faraday/acceptor"

	"github.com/lightninglabs/faraday/frdrpc"
	"github.com/lightninglabs/lndclient"
	"github.com/lightningnetwork/lnd/signal"
)

type Server struct {
	rpcServer *frdrpc.RPCServer
	acceptor  *acceptor.ChannelAcceptor
	quit      chan struct{}
}

// NewServer c
func NewServer(rpcServer *frdrpc.RPCServer,
	acceptor *acceptor.ChannelAcceptor) *Server {

	return &Server{
		rpcServer: rpcServer,
		acceptor:  acceptor,
		quit:      make(chan struct{}),
	}
}

func (s *Server) Start() error {
	// If we have a channel acceptor, start it.
	if s.acceptor != nil {
		if err := s.acceptor.Start(); err != nil {
			return err
		}
	}

	return s.rpcServer.Start()
}

func (s *Server) Stop() error {
	return s.rpcServer.Stop()
}

// Main is the real entry point for faraday. It is required to ensure that
// defers are properly executed when os.Exit() is called.
func Main() error {
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	// Connect to the full suite of lightning services offered by lnd's
	// subservers.
	client, err := lndclient.NewLndServices(&lndclient.LndServicesConfig{
		LndAddress:  config.RPCServer,
		Network:     lndclient.Network(config.network),
		MacaroonDir: config.MacaroonDir,
		TLSPath:     config.TLSCertPath,
		// Use the default lnd version check which checks for version
		// 10.1 and requires all build tags.
		CheckVersion: nil,
	})
	if err != nil {
		return fmt.Errorf("cannot connect to lightning services: %v",
			err)
	}
	defer client.Close()

	// Instantiate the faraday gRPC server.
	rpcServer := frdrpc.NewRPCServer(
		&frdrpc.Config{
			Lnd:        client.LndServices,
			RPCListen:  config.RPCListen,
			RESTListen: config.RESTListen,
			CORSOrigin: config.CORSOrigin,
		},
	)

	// If we have a channel acceptor set, then we create one.
	var chanAccept *acceptor.ChannelAcceptor

	if len(config.rejectPeers) > 0 || len(config.acceptPeers) > 0 {
		chanAccept = &acceptor.ChannelAcceptor{
			Accept:      client.Client.ChannelAcceptor,
			AcceptPeers: config.acceptPeers,
			RejectPeers: config.rejectPeers,
		}
	}

	server := NewServer(rpcServer, chanAccept)

	// Catch intercept signals, then start the server.
	signal.Intercept()
	if err := server.Start(); err != nil {
		return err
	}

	// Run until the user terminates.
	<-signal.ShutdownChannel()
	log.Infof("Received shutdown signal.")

	if err := server.Stop(); err != nil {
		return err
	}

	return nil
}
