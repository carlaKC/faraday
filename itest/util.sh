#!/bin/bash

function waitnoerror() {
        for i in {1..30}; do $@ && return; sleep 1; done
        echo "timeout"
        exit 1
}

function start_bitcoind() {
        echo "Starting bitcoind"
        bitcoind -regtest -txindex -rpcuser=devuser -rpcpassword=devpass \
        -zmqpubrawblock=tcp://0.0.0.0:29332 -zmqpubrawtx=tcp://0.0.0.0:29333 &

        BTCD_PID=$!

        BTCCTL="bitcoin-cli -regtest -rpcuser=devuser -rpcpassword=devpass"

        # Wait for btcd startup
        waitnoerror $BTCCTL getblockchaininfo
}

function start_lnds() {
        echo "Starting lnd"
        lnd --bitcoin.active --bitcoin.node=bitcoind --bitcoin.regtest --bitcoind.rpcuser=devuser \
        --bitcoind.zmqpubrawblock=tcp://localhost:29332 --bitcoind.zmqpubrawtx=tcp://localhost:29333 \
        --bitcoind.rpcpass=devpass --noseedbackup --nobootstrap --lnddir=lnd-alice \
        -d trace | awk '{ print "[lnd-alice] " $0; }' &

        LND_SERVER_PID=$!

        lnd --bitcoin.active --bitcoin.node=bitcoind --bitcoin.regtest --bitcoind.rpcuser=devuser \
        --bitcoind.rpcpass=devpass --noseedbackup --nobootstrap  --rpclisten=localhost:10002 \
        --bitcoind.zmqpubrawblock=tcp://localhost:29332 --bitcoind.zmqpubrawtx=tcp://localhost:29333 \
        --listen=localhost:10012 --restlisten=localhost:8002 \
        --lnddir=lnd-bob | awk '{ print "[lnd-bob] " $0; }' &

        LND_CLIENT_PID=$!

        LNCLI_SERVER="lncli --network regtest --lnddir lnd-alice"
        LNCLI_CLIENT="lncli --network regtest --lnddir lnd-bob --rpcserver=localhost:10002"

        waitnoerror $LNCLI_SERVER getinfo
        waitnoerror $LNCLI_CLIENT getinfo
}

function stop_all() {
        $LNCLI_CLIENT stop
        $LNCLI_SERVER stop
        $BTCCTL stop

        wait $BTCD_PID
        wait $LND_CLIENT_PID
        wait $LND_SERVER_PID
}