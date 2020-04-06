// Package faraday contains the main function for faraday.
package faraday

import (
	"fmt"

	"github.com/lightninglabs/faraday/frdrpc"
	"github.com/lightninglabs/loop/lndclient"
	"github.com/lightningnetwork/lnd/signal"
)

// Main is the real entry point for faraday. It is required to ensure that
// defers are properly executed when os.Exit() is called.
func Main() error {
	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	// NewBasicClient get a lightning rpc client with
	client, err := lndclient.NewBasicClient(
		config.RPCServer,
		config.TLSCertPath,
		config.MacaroonDir,
		config.network,
		lndclient.MacFilename(config.MacaroonFile),
	)
	if err != nil {
		return fmt.Errorf("cannot connect to lightning client: %v",
			err)
	}

	services, err := lndclient.NewLndServices(
		config.RPCServer,
		config.network,
		config.MacaroonDir,
		config.TLSCertPath,
	)
	if err != nil {
		return fmt.Errorf("cannot connect to services: %v", err)
	}

	// Instantiate the faraday gRPC server.
	server := frdrpc.NewRPCServer(
		&frdrpc.Config{
			LightningClient: client,
			GRPCServices:    services,
			RPCListen:       config.RPCListen,
		},
	)

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
