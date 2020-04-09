package routing

import (
	"time"

	"github.com/lightningnetwork/lnd/lnwire"
)

type Monitor interface {
	FailureRatio(startTime, endTime time.Time,
		channel lnwire.ShortChannelID, includeOurs bool) (float64,
		error)

	IncomingFailureRatio(startTime, endTime time.Time,
		channel lnwire.ShortChannelID, includeReceives bool) (float64,
		error)

	OutgoingFailureRatio(startTime, endTime time.Time,
		channel lnwire.ShortChannelID, includeSends bool) (float64,
		error)
}
