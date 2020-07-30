package accounting

import (
	"context"

	"github.com/lightninglabs/faraday/utils"
	"github.com/lightninglabs/lndclient"
)

// OnChainReport produces a report of our on chain activity for a period using
// live price data. Note that this report relies on transactions returned by
// GetTransactions in lnd. If a transaction is not included in this response
// (eg, a remote party opening a channel to us), it will not be included.
func OnChainReport(ctx context.Context, cfg *OnChainConfig) (Report, error) {
	// Retrieve a function which can be used to query individual prices,
	// or a no-op function if we do not want prices.
	getPrice, err := getConversion(
		ctx, cfg.StartTime, cfg.EndTime, cfg.DisableFiat,
		cfg.Granularity,
	)
	if err != nil {
		return nil, err
	}

	return onChainReportWithPrices(cfg, getPrice)
}

// onChainReportWithPrices produces off chain reports using the getPrice
// function provided. This allows testing of our report creation without calling
// the actual price API.
func onChainReportWithPrices(cfg *OnChainConfig, getPrice usdPrice) (Report,
	error) {

	// Create an info struct to hold all the elements we need.
	info := onChainInformation{
		priceFunc:      getPrice,
		feeFunc:        cfg.GetFee,
		openChannels:   make(map[string]lndclient.ChannelInfo),
		sweeps:         make(map[string]bool),
		closedChannels: make(map[string]lndclient.ClosedChannel),
		channelCloses:  make(map[string]lndclient.ClosedChannel),
	}

	onChainTxns, err := cfg.OnChainTransactions()
	if err != nil {
		return nil, err
	}

	// Filter our on chain transactions by start and end time. If we have
	// no on chain transactions over this period, we can return early.
	info.txns = filterOnChain(cfg.StartTime, cfg.EndTime, onChainTxns)
	if len(info.txns) == 0 {
		return Report{}, nil
	}

	// Get our opened channels and create a map of closing txid to the
	// channel entry. This will be used to separate channel opens out from
	// other on chain transactions.
	openRPCChannels, err := cfg.OpenChannels()
	if err != nil {
		return nil, err
	}

	for _, channel := range openRPCChannels {
		outpoint, err := utils.GetOutPointFromString(
			channel.ChannelPoint,
		)
		if err != nil {
			return nil, err
		}

		// Add the channel to our map, keyed by txid.
		info.openChannels[outpoint.Hash.String()] = channel
	}

	// Get our closed channels and create a map of closing txid to closed
	// channel. This will be used to separate out channel closes from other
	// on chain transactions.
	closedRPCChannels, err := cfg.ClosedChannels()
	if err != nil {
		return nil, err
	}

	// Add our already closed channels open and closed transactions to our
	// on chain info so that we will be able to detect channels that were
	// opened and closed within our period.
	for _, closedChannel := range closedRPCChannels {
		info.channelCloses[closedChannel.ClosingTxHash] = closedChannel

		outpoint, err := utils.GetOutPointFromString(
			closedChannel.ChannelPoint,
		)
		if err != nil {
			return nil, err
		}
		info.channelCloses[outpoint.Hash.String()] = closedChannel
	}

	// Finally, get our list of known sweeps from lnd so that we can
	// identify them separately to other on chain transactions.
	sweeps, err := cfg.ListSweeps()
	if err != nil {
		return nil, err
	}

	for _, sweep := range sweeps {
		info.sweeps[sweep] = true
	}

	return onChainReport(info)
}

// onChainInformation contains all the information we require to produce an
// on chain report.
type onChainInformation struct {
	txns           []lndclient.Transaction
	priceFunc      usdPrice
	feeFunc        getFeeFunc
	sweeps         map[string]bool
	openChannels   map[string]lndclient.ChannelInfo
	closedChannels map[string]lndclient.ClosedChannel
	channelCloses  map[string]lndclient.ClosedChannel
}

// onChainReport produces an on chain transaction report.
func onChainReport(info onChainInformation) (
	Report, error) {

	var report Report

	for _, txn := range info.txns {
		// First, we check whether our transaction is a channel close,
		// because channel closes may report as having a zero amount (in
		// the case of a force close) and are expected to have a zero
		// fee amount because the wallet does not account for fees that
		// are taken from the input we are spending.
		channelClose, ok := info.channelCloses[txn.TxHash]
		if ok {
			entries, err := closedChannelEntries(
				channelClose, txn, info.feeFunc, info.priceFunc,
			)
			if err != nil {
				return nil, err
			}

			report = append(report, entries...)
			continue
		}

		// Next, we check whether our transaction is a sweep. In this
		// case we also expect to have zero fees, because sweeps use
		// their existing inputs for fees.
		if info.sweeps[txn.TxHash] {
			entries, err := sweepEntries(
				txn, info.feeFunc, info.priceFunc,
			)
			if err != nil {
				return nil, err
			}

			report = append(report, entries...)
			continue
		}

		// If the transaction is in our set of currently open channels,
		// we just need an open channel entry for it.
		openChannel, ok := info.openChannels[txn.TxHash]
		if ok {
			entries, err := channelOpenEntries(
				openChannel, txn, info.priceFunc,
			)
			if err != nil {
				return nil, err
			}
			report = append(report, entries...)
			continue
		}

		// If the transaction is a channel opening transaction for one
		// of our already closed channels, we need to reconstruct a
		// channel open from our close summary.
		channelOpen, ok := info.closedChannels[txn.TxHash]
		if ok {
			entries, err := openChannelFromCloseSummary(
				channelOpen, txn, info.priceFunc,
			)
			if err != nil {
				return nil, err
			}

			report = append(report, entries...)
			continue
		}

		entries, err := onChainEntries(txn, info.priceFunc)
		if err != nil {
			return nil, err
		}
		report = append(report, entries...)
	}

	return report, nil
}
