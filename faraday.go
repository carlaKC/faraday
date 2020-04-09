// Package faraday contains the main function for faraday.
package faraday

import (
	"context"
	"fmt"

	"github.com/lightninglabs/faraday/frdrpc"
	"github.com/lightninglabs/faraday/routing"
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

	routingMonitor := routing.NewRoutingMonitor(
		routing.Config{
			SubscribeEvents: func() (i <-chan interface{},
				errors <-chan error, err error) {

				// TODO(carla): figure out where this should live?
				// I think the context should belong to us on router
				// level so that we can cancel it?
				return services.Router.SubscribeHtlcEvents(context.Background())
			}},
	)

	if err := routingMonitor.Start(); err != nil {
		return err
	}

	// Instantiate the faraday gRPC server.
	server := frdrpc.NewRPCServer(
		&frdrpc.Config{
			LightningClient: client,
			GRPCServices:    services,
			RoutingMonitor:  routingMonitor,
			RPCListen:       config.RPCListen,
		},
	)

	if err := server.Start(); err != nil {
		return err
	}

	// Run until the user terminates.
	<-signal.ShutdownChannel()
	log.Infof("Received shutdown signal.")

	if err := routingMonitor.Stop(); err != nil {
		return err
	}

	if err := server.Stop(); err != nil {
		return err
	}

	return nil
}
