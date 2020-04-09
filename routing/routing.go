// Package routing is a subsystem of faraday which gathers routing information
// from lnd.
package routing

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/lightningnetwork/lnd/lnwire"
)

var (
	// errNotHtlcEvent is returned when we get an event type we do not expect
	// from the stream.
	errNotHtlcEvent = errors.New("event from stream not htlc event")

	// errEventAlreadyIndexed is returned when we have already processed an
	// event with the key provided.
	errEventAlreadyIndexed = errors.New("event with index already exists")

	// errIndexedNonForward is returned when we lookup the corresponding
	// forward for a htlc settle/failure and find a htlc event of another
	// type at that index in our set of events.
	errIndexedNonForward = errors.New("event that is not an forwarding " +
		"event at index")
)

// RoutingConfig contains all the functions the routing monitor requires to
// track lnd routing events.
type Config struct {
	SubscribeEvents func() (<-chan interface{}, <-chan error, error)
}

type routingMonitor struct {
	cfg Config

	quit chan struct{}

	// HtlcEvents is a time series set of htlc events.
	HtlcEvents []htlcEvent

	// ForwardIndex maps a unique event key to its index in the HtlcEvents
	// slice for fast querying. This index is only queried for forwarding
	// events that can be subsequently failed or settled.
	ForwardIndex map[string]int

	// failureRatioRequests serves requests for the lifespan of channels.
	failureRatioRequests chan *failureRatioRequest

	wg sync.WaitGroup
}

// NewRoutingMonitor returns a routing monitor.
func NewRoutingMonitor(cfg Config) *routingMonitor {
	return &routingMonitor{
		cfg:                  cfg,
		quit:                 make(chan struct{}),
		HtlcEvents:           nil,
		ForwardIndex:         make(map[string]int),
		failureRatioRequests: make(chan *failureRatioRequest),
	}
}

// Start starts monitoring our node's routing activity.
func (r *routingMonitor) Start() error {
	events, errChan, err := r.cfg.SubscribeEvents()
	if err != nil {
		return err
	}

	r.wg.Add(1)
	go r.monitor(events, errChan)

	return nil
}

// Stop signals all active goroutines to terminate  and waits for them to exit.
func (r *routingMonitor) Stop() error {
	close(r.quit)
	r.wg.Wait()

	return nil
}

// monitor is the main events bus for the routing monitor. It consumes htlc
// routing and channel open/closure events to track our node's routing activity.
func (r *routingMonitor) monitor(eventsChan <-chan interface{},
	errChan <-chan error) {

	defer r.wg.Done()

	for {
		select {
		case event := <-eventsChan:
			if err := r.processHtlcEvent(event); err != nil {
				log.Warnf("process htlc error: %v", err)
				return
			}

		// If our htlc event subscription fails, exit.
		case err := <-errChan:
			log.Warnf("htlc subscription error: %v", err)
			return

		// If we get a request to calculate failure ratio, we assemble
		// a set of filters to get the correct htlc set, and return
		// the failure ratio calculated.
		case req := <-r.failureRatioRequests:
			filters := []filterHtlcEvent{
				filterStartTime(req.startTime),
				filterEndTime(req.endTime),
			}

			// If we want to exclude outgoing htlcs from the set of
			// htlcs we get failure ratios for, we add a filter
			// to remove all outgoing htlcs/failures.
			if req.excludeOutgoing {
				filters = append(filters, filterOutgoing())
			}

			// If we want to exclude incoming htlcs from the set of
			// htlcs we get failure ratios for, we add a filter
			// to remove all incoming htlcs/failures.
			if req.excludeIncoming {
				filters = append(filters, filterIncoming())
			}

			// If we do not want to include our own sends in the
			// failure ratio calculation, we add a filter that will
			// remove all sends from our set of events.
			if !req.includeSends {
				filterSends := filterEventType(eventTypeSend)
				filters = append(filters, filterSends)
			}

			// If we do not want to include our own receives in the
			// failure ratio calculation, we add a filter that will
			// remove all receives from our set of events.
			if !req.includeReceives {
				filterReceives := filterEventType(
					eventTypeReceive,
				)
				filters = append(filters, filterReceives)
			}

			events := r.filterEvents(filters...)
			ratio, err := failureRatio(events)
			req.responseChan <- failureRateResponse{
				failureRatio: ratio,
				err:          err,
			}

		// Exit if the routing monitor receives the signal to
		// shutdown.
		case <-r.quit:
			return
		}
	}
}

// failureRatio calculates the ratio of htlc failures to total htlcs over a set
// of htlc events. This function expects only resolved htlc events to be passed
// in.
func failureRatio(events []htlcEvent) (float64, error) {
	log.Debug("calculating failure ratio for: %v events", len(events))

	if len(events) == 0 {
		return 0, nil
	}

	var totalFailures int

	for _, event := range events {
		switch e := event.(type) {
		case *htlcForward:
			if e.resolution == nil {
				return 0, fmt.Errorf("failureRatio received "+
					"unresolved htlc event: %v", e.key)
			}

			// If the htlc was not settled, it failed, so we
			// increment our
			if !e.resolution.settled {
				totalFailures++
			}

		case *linkFailure:
			totalFailures++
		}
	}

	log.Debug("failure ratio: %v / %v", totalFailures, len(events))

	return float64(totalFailures) / float64(len(events)), nil
}

func eventKey(chanIn, indexIn, chanOut, indexOut uint64) string {
	return fmt.Sprintf("%v:%v-%v:%v", chanIn, indexIn, chanOut, indexOut)
}

// TODO(carla): make this routerrpc naive.
func (r *routingMonitor) processHtlcEvent(event interface{}) error {
	htlcEvent, ok := event.(routerrpc.HtlcEvent)
	if !ok {
		return errNotHtlcEvent
	}

	key := eventKey(
		htlcEvent.IncomingChannelId, htlcEvent.IncomingHtlcId,
		htlcEvent.OutgoingChannelId, htlcEvent.OutgoingHtlcId,
	)

	timestamp := time.Unix(0, int64(htlcEvent.TimestampNs))

	var eventType eventType
	switch htlcEvent.EventType {
	case routerrpc.HtlcEvent_SEND:
		eventType = eventTypeSend

	case routerrpc.HtlcEvent_RECEIVE:
		eventType = eventTypeReceive

	case routerrpc.HtlcEvent_FORWARD:
		eventType = eventTypeForward
	}

	switch e := htlcEvent.Event.(type) {
	case *routerrpc.HtlcEvent_ForwardEvent:
		return r.addForwardEvent(key, timestamp, eventType, e)

	case *routerrpc.HtlcEvent_ForwardFailEvent:
		return r.addForwardingResolution(key, timestamp, false)

	case *routerrpc.HtlcEvent_LinkFailEvent:
		return r.addLinkFailure(key, timestamp, eventType, e)

	case *routerrpc.HtlcEvent_SettleEvent:
		if eventType == eventTypeReceive {
			return r.addReceive(key, timestamp)
		}

		return r.addForwardingResolution(key, timestamp, true)
	}

	return nil
}

func (r *routingMonitor) addForwardEvent(key string, timestamp time.Time,
	eventType eventType, event *routerrpc.HtlcEvent_ForwardEvent) error {

	// Sanity check that we do not already have an event with this key in
	// our index.
	_, ok := r.ForwardIndex[key]
	if ok {
		return errEventAlreadyIndexed
	}

	info := event.ForwardEvent.Info

	// Create a htlc forward event from the forwarding event we are
	// processing. We leave the resolution as nil because we have not yet
	// learned of this event's resolution.
	forward := &htlcForward{
		key:         key,
		forwardedAt: timestamp,
		eventType:   eventType,
		amount:      lnwire.MilliSatoshi(info.OutgoingAmtMsat),
		resolution:  nil,
	}

	// We only set our fee amount if we have a non-zero incoming amount,
	// otherwise we will have a negative fee value. This covers locally
	// initiated sends which have zero incoming amount/ time lock.
	if info.IncomingAmtMsat != 0 {
		forward.fees = lnwire.MilliSatoshi(
			info.IncomingAmtMsat - info.OutgoingAmtMsat,
		)
	}

	// TODO(carla): record timelock delta

	// Our event will be appended to our current slice of events. Its index
	// in this slice will be equal to the length of the array before we add
	// it, because last index = len-1.
	idx := len(r.HtlcEvents)
	r.HtlcEvents = append(r.HtlcEvents, forward)

	// Finally, save the index where we inserted this forward event to our
	// map so that we can quickly access it when we learn of its resolution.
	r.ForwardIndex[key] = idx

	return nil
}

func (r *routingMonitor) addForwardingResolution(key string,
	timestamp time.Time, settled bool) error {

	// Lookup the corresponding forwarding event in our index. We are not
	// guaranteed to find this event, because it is possible that the htlc
	// was forwarded before a lnd restart and failed back afterwards.
	index, ok := r.ForwardIndex[key]
	if !ok {
		log.Warnf("resolution: %v does not have a "+
			"matching forward", key)
		// TODO(carla): add this anyway, it does give us information?
		return nil
	}

	event := r.HtlcEvents[index]
	forwardEvent, ok := event.(*htlcForward)
	if !ok {
		return errIndexedNonForward
	}

	// Add a resolution to our forward with the timestamp of the resolution
	// event.
	forwardEvent.resolution = &resolution{
		resolvedAt: timestamp,
		settled:    settled,
	}

	// Finally, replace the forwarding event in our list of events with the
	// updated copy.
	r.HtlcEvents[index] = forwardEvent

	return nil
}

func (r *routingMonitor) addLinkFailure(key string, timestamp time.Time,
	eventType eventType, event *routerrpc.HtlcEvent_LinkFailEvent) error {
	linkFailure := &linkFailure{
		key:           key,
		failedAt:      timestamp,
		eventType:     eventType,
		incoming:      event.LinkFailEvent.Incoming,
		failureDetail: event.LinkFailEvent.FailureDetail,
		failureCode:   event.LinkFailEvent.WireFailure,
	}

	r.HtlcEvents = append(r.HtlcEvents, linkFailure)

	return nil
}

// addReceive when we add a htlc receive, we simply append the special case
// typed struct to the set of events we have.
func (r *routingMonitor) addReceive(key string, timestamp time.Time) error {
	r.HtlcEvents = append(
		r.HtlcEvents, &receive{key: key, settledAt: timestamp},
	)

	return nil
}
