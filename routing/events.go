package routing

import (
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/lightningnetwork/lnd/lnwire"
)

// eventType indicates the type of event the htlcs were involved in; a user
// initiated send, a recevie to our node, or a forward through our node.
type eventType int

const (
	eventTypeUnknown eventType = iota
	eventTypeForward
	eventTypeReceive
	eventTypeSend
)

type htlcEvent interface {
	Type() eventType
	Timestamp() time.Time
}

type htlcForward struct {
	key         string
	forwardedAt time.Time
	eventType   eventType

	amount lnwire.MilliSatoshi
	fees   lnwire.MilliSatoshi

	// The settle or fail that resolved the forwarded htlc. This field will
	// be nil if the htlc has not yet been resolved.
	resolution *resolution
}

type resolution struct {
	resolvedAt time.Time
	settled    bool
}

func (f *htlcForward) Type() eventType {
	return f.eventType
}

func (f *htlcForward) Timestamp() time.Time {
	return f.forwardedAt
}

type linkFailure struct {
	key       string
	failedAt  time.Time
	eventType eventType

	// Incoming indicates whether the link failure occurred on the incoming
	// or outgoing link.
	incoming bool

	// failureDetail is additional information that the link failure is
	// (optionally) enriched with.
	failureDetail routerrpc.FailureDetail

	// failureCode is the wire code for the failure.
	failureCode lnrpc.Failure_FailureCode
}

func (l *linkFailure) Type() eventType {
	return l.eventType
}

func (l *linkFailure) Timestamp() time.Time {
	return l.failedAt
}

type receive struct {
	key       string
	settledAt time.Time
}

func (s *receive) Type() eventType {
	return eventTypeReceive
}

func (s *receive) Timestamp() time.Time {
	return s.settledAt
}
