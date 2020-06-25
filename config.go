package faraday

import (
	"fmt"
	"strings"
	"time"

	"github.com/lightningnetwork/lnd/routing/route"

	"github.com/jessevdk/go-flags"
	"github.com/lightningnetwork/lnd/build"
)

const (
	defaultRPCPort        = "10009"
	defaultRPCHostPort    = "localhost:" + defaultRPCPort
	defaultNetwork        = "mainnet"
	defaultMinimumMonitor = time.Hour * 24 * 7 * 4 // four weeks in hours
	defaultDebugLevel     = "info"
	defaultRPCListen      = "localhost:8465"
)

type Config struct {
	// RPCServer is host:port that lnd's RPC server is listening on.
	RPCServer string `long:"rpcserver" description:"host:port that LND is listening for RPC connections on"`

	// MacaroonDir is the directory containing macaroons.
	MacaroonDir string `long:"macaroondir" description:"Dir containing macaroons"`

	// TLSCertPath is the path to the tls cert that faraday should use.
	TLSCertPath string `long:"tlscertpath" description:"Path to TLS cert"`

	// TestNet is set to true when running on testnet.
	TestNet bool `long:"testnet" description:"Use the testnet network"`

	// Simnet is set to true when using btcd's simnet.
	Simnet bool `long:"simnet" description:"Use simnet"`

	// Simnet is set to true when using bitcoind's regtest.
	Regtest bool `long:"regtest" description:"Use regtest"`

	// MinimumMonitored is the minimum amount of time that a channel must be monitored for before we consider it for termination.
	MinimumMonitored time.Duration `long:"min_monitored" description:"The minimum amount of time that a channel must be monitored for before recommending termination. Valid time units are {s, m, h}."`

	// network is a string containing the network we're running on.
	network string

	// DebugLevel is a string defining the log level for the service either
	// for all subsystems the same or individual level by subsystem.
	DebugLevel string `long:"debuglevel" description:"Debug level for faraday and its subsystems."`

	// RPCListen is the listen address for the faraday rpc server.
	RPCListen string `long:"rpclisten" description:"Address to listen on for gRPC clients."`

	// RESTListen is the listen address for the faraday REST server.
	RESTListen string `long:"restlisten" description:"Address to listen on for REST clients. If not specified, no REST listener will be started."`

	// CORSOrigin specifies the CORS header that should be set on REST responses. No header is added if the value is empty.
	CORSOrigin string `long:"corsorigin" description:"The value to send in the Access-Control-Allow-Origin header. Header will be omitted if empty."`

	// PeerAcceptList is a set of peers that we will accept channel opens from, mutually exclusive with PeerRejectList.
	PeerAcceptList string `long:"peer_accept" description:"The set of peer pubkeys (separated by commas) that we will accept channel opens from. Be mindful that if set, your node will *only* accept channels from these peers. Note this cannot be used if peer_reject is set."`

	// PeerRejectList a set of peers we won't accept channel opens from, mutually exclusive with PeerAcceptList.
	PeerRejectList string `long:"peer_reject" description:"A list of peer pubkeys (separated by commas) that we will reject channel opens from. Note this cannot be used if peer_accept is set."`

	// acceptPeers contains the peers from PeerAcceptList.
	acceptPeers []route.Vertex

	// rejectPeers contains the peers from PeerRejectList.
	rejectPeers []route.Vertex
}

// DefaultConfig returns all default values for the Config struct.
func DefaultConfig() Config {
	return Config{
		RPCServer:        defaultRPCHostPort,
		network:          defaultNetwork,
		MinimumMonitored: defaultMinimumMonitor,
		DebugLevel:       defaultDebugLevel,
		RPCListen:        defaultRPCListen,
	}
}

// LoadConfig starts with a skeleton default config, and reads in user provided
// configuration from the command line. It does not provide a full set of
// defaults or validate user input because validation and sensible default
// setting are performed by the lndclient package.
func LoadConfig() (*Config, error) {
	// Start with a default config.
	config := DefaultConfig()

	// Parse command line options to obtain user specified values.
	if _, err := flags.Parse(&config); err != nil {
		return nil, err
	}

	var netCount int
	if config.TestNet {
		config.network = "testnet"
		netCount++
	}
	if config.Regtest {
		config.network = "regtest"
		netCount++
	}
	if config.Simnet {
		config.network = "simnet"
		netCount++
	}

	if netCount > 1 {
		return nil, fmt.Errorf("do not specify more than one network flag")
	}

	if err := build.ParseAndSetDebugLevels(config.DebugLevel, logWriter); err != nil {
		return nil, err
	}

	if config.PeerAcceptList != "" {
		accept := strings.Split(config.PeerAcceptList, ",")

		for _, peer := range accept {
			vertex, err := route.NewVertexFromStr(peer)
			if err != nil {
				return nil, fmt.Errorf("could not parse "+
					"accept peer pubkey: %v: %v", peer, err)
			}

			config.acceptPeers = append(config.acceptPeers, vertex)
		}
	}

	if config.PeerRejectList != "" {
		reject := strings.Split(config.PeerRejectList, ",")

		for _, peer := range reject {
			vertex, err := route.NewVertexFromStr(peer)
			if err != nil {
				return nil, fmt.Errorf("could not parse "+
					"reject peer pubkey: %v: %v", peer, err)
			}

			config.rejectPeers = append(config.rejectPeers, vertex)
		}
	}

	// Do not allow setting of accept and reject lists.
	if len(config.acceptPeers) > 0 && len(config.rejectPeers) > 0 {
		return nil, fmt.Errorf("cannot set %v peer_reject and "+
			"%v peer_accept, pick one", len(config.acceptPeers),
			len(config.rejectPeers))
	}

	return &config, nil
}
