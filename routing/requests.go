package routing

import (
	"errors"
	"time"

	"github.com/lightningnetwork/lnd/lnwire"
)

var (
	errShuttingDown = errors.New("routing monitor shutting down")

	errStartAfterEnd = errors.New("start time is after end time")
)

// failureRatioRequest queries the routing monitor for the ratio of htlc
// failures to total htlcs. Various parameters can be used to which areas we
// calculate this failure ratio for:
// TODO
// excludeOutgoing: we want the failure
// excludeIncoming:
// excludeOutgoing + includeReceives:
// excludeIncoming + includeSends:
type failureRatioRequest struct {
	startTime time.Time
	endTime   time.Time

	// If channel ID is set, the query will return failure rate of only that
	// channel.
	// TODO(carla): implement this
	channel lnwire.ShortChannelID

	// excludeOutgoing is set to true if we only want to get the failure
	// ratio of htlcs arriving at our node. This will exclude our own sends
	// and forwards that failed on our outgoing link from our calculations.
	excludeOutgoing bool

	// excludeIncoming is set to true if we only want to get the failure
	// ratio of htlcs leaving our node. This will exclude failures that
	// occurred on our incoming link (forwards or receives) from our failure
	// count, and exclude receives from our total.
	excludeIncoming bool

	// includeSends will include our sends in the failure rate query. They
	// are excluded by default.
	includeSends bool

	// includeReceives will include our receives in the failure rate query.
	// They are excluded by default.
	includeReceives bool

	responseChan chan failureRateResponse
}

type failureRateResponse struct {
	failureRatio float64
	err          error
}

// FailureRatio returns the ratio of failures of our incoming and outgoing
// htlcs to total htlcs received and forwarded onwards. If includeOwn is set to
// true, our own sends and receives will be included in this ratio.
func (r *routingMonitor) FailureRatio(startTime, endTime time.Time,
	channel lnwire.ShortChannelID, includeOurs bool) (float64, error) {

	// We set excludeIncoming and excludeOUtgoing to false, because we want
	// failure rate for all of our htlcs. We include sends and receives
	// based on the user's input.
	return r.getFailureRatio(
		startTime, endTime, channel, false, false, includeOurs,
		includeOurs,
	)
}

// OutgoingFailureRatio returns the failure ratio of failures that occurred
// for htlcs leaving our node on the outgoing link to total number of htlcs
// that should have left our link over a time period. If includeSends is set to
// true, our own sends will be included in this calculation.
func (r *routingMonitor) OutgoingFailureRatio(startTime, endTime time.Time,
	channel lnwire.ShortChannelID, includeSends bool) (float64, error) {

	// We set excludeIncoming to true and excludeOUtgoing to false, because
	// we only want incoming failure rate, for this query. We set include
	// sends to receives (we won't have any outgoing htlcs that are
	// receives, but set it false nonetheless) and include receives based
	// on user request.
	return r.getFailureRatio(
		startTime, endTime, channel, false, true, includeSends,
		false,
	)
}

// OutgoingFailureRatio returns the failure ratio of failures that occurred
// for htlcs arriving our node on the incoming link to total number of htlcs
// that arrived at our node over a time period. If includeReceives is set to
// true, our own receives will be included in this calculation.
func (r *routingMonitor) IncomingFailureRatio(startTime, endTime time.Time,
	channel lnwire.ShortChannelID, includeReceives bool) (float64, error) {

	// We set excludeOutgoing to true and excludeIncoming to false, because
	// we only want incoming failure rate, for this query. We set include
	// sends to false (we won't have any incoming htlcs that are sends, but
	// set it false nonetheless) and include receives based on user request.
	return r.getFailureRatio(
		startTime, endTime, channel, true, false, false,
		includeReceives,
	)
}

// getFailureRatio creates a request for a failure ratio with the parameters
// provided.
func (r *routingMonitor) getFailureRatio(startTime, endTime time.Time,
	channel lnwire.ShortChannelID, excludeOutgoing, excludeIncoming,
	includeSends, includeReceives bool) (float64,
	error) {

	request := &failureRatioRequest{
		startTime:       startTime,
		endTime:         endTime,
		channel:         channel,
		excludeOutgoing: excludeOutgoing,
		excludeIncoming: excludeIncoming,
		includeSends:    includeSends,
		includeReceives: includeReceives,
		responseChan:    make(chan failureRateResponse),
	}

	// If the end time provided is zero, progress it to the current time.
	if endTime.IsZero() {
		request.endTime = time.Now()
	}

	// If start time is after end time, after we have progressed our
	// end time to the present if it was zero, fail the query.
	if startTime.After(endTime) {
		return 0, errStartAfterEnd
	}

	// Send the request for failure ratio to our main events bus.
	select {
	case r.failureRatioRequests <- request:

	case <-r.quit:
		return 0, errShuttingDown
	}

	// Return the response we receive on the response channel or exit early
	// if the store is instructed to exit.
	select {
	case resp := <-request.responseChan:
		return resp.failureRatio, resp.err

	case <-r.quit:
		return 0, errShuttingDown
	}
}
