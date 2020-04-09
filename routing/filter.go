package routing

import (
	"time"

	"github.com/lightningnetwork/lnd/lnwire"
)

func (r *routingMonitor) filterEvents(filters ...filterHtlcEvent) []htlcEvent {
	var results []htlcEvent

	// Iterate through our set of events.
	for _, event := range r.HtlcEvents {
		include := true

		// Iterate through our set of filters. If any of the filters
		// return true, we skip the event because it does not fit our
		// search criteria. Once one filter has returned true, we do
		// not need to evaluate any more.
		for _, filter := range filters {
			if filter(event) {
				include = false
				break
			}
		}

		if include {
			results = append(results, event)
		}
	}

	return results
}

// filterHtlcEvent is a function which can be used to filter a set of htlc
// events. If it returns true, the htlc event should not be included in our
// set of returned events.
type filterHtlcEvent func(h htlcEvent) bool

// filterStartTime filters out events which are before a given start time.
func filterStartTime(startTime time.Time) func(h htlcEvent) bool {
	return func(h htlcEvent) bool {
		return h.Timestamp().Before(startTime)
	}
}

// filterEndTime filters out events which are after a given end time.
func filterEndTime(endTime time.Time) func(h htlcEvent) bool {
	return func(h htlcEvent) bool {
		return h.Timestamp().After(endTime)
	}
}

// filterEventType filters out events of a give type.
func filterEventType(eventType eventType) func(h htlcEvent) bool {
	return func(h htlcEvent) bool {
		return h.Type() == eventType
	}
}

func filterUnresolved() func(h htlcEvent) bool {
	return func(h htlcEvent) bool {
		// If the event is not a htlc forward, it is a receive or a link
		// error, both of which are fully resolved.
		fwd, ok := h.(*htlcForward)
		if !ok {
			return false
		}

		// If the forward's resolution is nil, we should filter it out.
		return fwd.resolution == nil
	}
}

// filterChannel filters out all htlcs that did not occur in the channel
// provided. If the channel id provided is zero, nothing will be filtered out.
func filterChannel(id lnwire.ShortChannelID) func(h htlcEvent) bool {
	return func(h htlcEvent) bool {
		chanID := id.ToUint64()

		// If a zero channel ID was provided, we do not want to filter
		// anything.
		if chanID == 0 {
			return false
		}

		// TODO(carla): need to filter all channels that do not equal
		// this one
		return false
	}
}

// filterIncoming filters out failures that occurred on our incoming link
// and receives (which arrive on an incoming link and do not leave our node).
func filterIncoming() func(h htlcEvent) bool {
	return func(h htlcEvent) bool {
		switch event := h.(type) {
		case *receive:
			return true

		case *linkFailure:
			return event.incoming
		}

		return false
	}
}

// filterOutgoing filters out failures that occurred on our outgoing link.
func filterOutgoing() func(h htlcEvent) bool {
	return func(h htlcEvent) bool {

		switch event := h.(type) {
		// If we have a htlc forward that is a send, this htlc never
		// arrived at our node on the incoming link, so we filter it out.
		case *htlcForward:
			return event.eventType == eventTypeSend

		// If the link failure is not incoming, it occurred on our
		// outgoing link, so we filter it out.
		case *linkFailure:
			return !event.incoming
		}

		return false
	}
}

/*
Things that we want to have:
- average success/ failure rate
- average bandwidth usage over time
- time consumed by failed/settled htlcs?
- tracking of most common errors? bandwidth
- how much we use for our own use
*/
