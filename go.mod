module github.com/lightninglabs/faraday

require (
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btclog v0.0.0-20170628155309-84c8d2346e9f
	github.com/golang/protobuf v1.3.3
	github.com/jessevdk/go-flags v1.4.0
	github.com/lightninglabs/loop v0.2.4-alpha
	github.com/lightninglabs/protobuf-hex-display v1.3.3-0.20191212020323-b444784ce75d
	github.com/lightningnetwork/lnd v0.9.0-beta-rc3.0.20200121213302-a2977c4438b5
	github.com/urfave/cli v1.20.0
	google.golang.org/genproto v0.0.0-20190927181202-20e1ac93f88c
	google.golang.org/grpc v1.27.0
)

replace github.com/lightninglabs/loop => /Users/carla/go/src/github.com/lightninglabs/loop

replace github.com/lightningnetwork/lnd => /Users/carla/go/src/github.com/lightningnetwork/lnd

go 1.13
