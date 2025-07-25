package ws

import (
	"orly.dev/pkg/encoders/envelopes/closeenvelope"
	"orly.dev/pkg/encoders/envelopes/countenvelope"
	"orly.dev/pkg/encoders/envelopes/reqenvelope"
	"orly.dev/pkg/encoders/event"
	"orly.dev/pkg/encoders/filters"
	"orly.dev/pkg/encoders/subscription"
	"orly.dev/pkg/utils/chk"
	"orly.dev/pkg/utils/context"
	"orly.dev/pkg/utils/errorf"
	"strconv"
	"sync"
	"sync/atomic"
)

// Subscription is a client interface for a subscription (what REQ turns into
// after EOSE).
type Subscription struct {
	label   string
	counter int

	Relay   *Client
	Filters *filters.T

	// for this to be treated as a COUNT and not a REQ this must be set
	countResult chan int

	// The Events channel emits all EVENTs that come in a Subscription will be
	// closed when the subscription ends
	Events event.C
	mu     sync.Mutex

	// The EndOfStoredEvents channel is closed when an EOSE comes for that
	// subscription
	EndOfStoredEvents chan struct{}

	// The ClosedReason channel emits the reason when a CLOSED message is
	// received
	ClosedReason chan string

	// Context will be .Done() when the subscription ends
	Context context.T

	live   atomic.Bool
	eosed  atomic.Bool
	closed atomic.Bool
	cancel context.F

	// This keeps track of the events we've received before the EOSE that we
	// must dispatch before closing the EndOfStoredEvents channel
	storedwg sync.WaitGroup
}

// EventMessage is an event, with the associated relay URL attached.
type EventMessage struct {
	Event event.E
	Relay string
}

// SubscriptionOption is the type of the argument passed for that. Some examples
// are WithLabel.
type SubscriptionOption interface {
	IsSubscriptionOption()
}

// WithLabel puts a label on the subscription (it is prepended to the automatic
// id) that is sent to relays.
type WithLabel string

func (_ WithLabel) IsSubscriptionOption() {}

var _ SubscriptionOption = (WithLabel)("")

// GetID return the Nostr subscription ID as given to the Client it is a
// concatenation of the label and a serial number.
func (sub *Subscription) GetID() (id *subscription.Id) {
	var err error
	if id, err = subscription.NewId(sub.label + ":" + strconv.Itoa(sub.counter)); chk.E(err) {
		return
	}
	return
}

func (sub *Subscription) start() {
	<-sub.Context.Done()
	// the subscription ends once the context is canceled (if not already)
	sub.Unsub() // this will set sub.live to false

	// do this so we don't have the possibility of closing the Events channel
	// and then trying to send to it
	sub.mu.Lock()
	close(sub.Events)
	sub.mu.Unlock()
}

func (sub *Subscription) dispatchEvent(evt *event.E) {
	added := false
	if !sub.eosed.Load() {
		sub.storedwg.Add(1)
		added = true
	}

	go func() {
		sub.mu.Lock()
		defer sub.mu.Unlock()

		if sub.live.Load() {
			select {
			case sub.Events <- evt:
			case <-sub.Context.Done():
			}
		}

		if added {
			sub.storedwg.Done()
		}
	}()
}

func (sub *Subscription) dispatchEose() {
	if sub.eosed.CompareAndSwap(false, true) {
		go func() {
			sub.storedwg.Wait()
			sub.EndOfStoredEvents <- struct{}{}
		}()
	}
}

func (sub *Subscription) dispatchClosed(reason string) {
	if sub.closed.CompareAndSwap(false, true) {
		go func() {
			sub.ClosedReason <- reason
		}()
	}
}

// Unsub closes the subscription, sending "CLOSE" to relay as in NIP-01. Unsub()
// also closes the channel sub.Events and makes a new one.
func (sub *Subscription) Unsub() {
	// cancel the context (if it's not canceled already)
	sub.cancel()
	// mark the subscription as closed and send a CLOSE to the relay (naïve
	// sync.Once implementation)
	if sub.live.CompareAndSwap(true, false) {
		sub.Close()
	}
	// remove subscription from our map
	sub.Relay.Subscriptions.Delete(sub.GetID().String())
}

// Close just sends a CLOSE message. You probably want Unsub() instead.
func (sub *Subscription) Close() {
	if sub.Relay.IsConnected() {
		id := sub.GetID()
		closeMsg := closeenvelope.NewFrom(id)
		var b []byte
		b = closeMsg.Marshal(nil)
		<-sub.Relay.Write(b)
	}
}

// Sub sets sub.Filters and then calls sub.Fire(ctx). The subscription will be
// closed if the context expires.
func (sub *Subscription) Sub(_ context.T, ff *filters.T) {
	sub.Filters = ff
	sub.Fire()
}

// Fire sends the "REQ" command to the relay.
func (sub *Subscription) Fire() (err error) {
	id := sub.GetID()

	var b []byte
	if sub.countResult == nil {
		b = reqenvelope.NewFrom(id, sub.Filters).Marshal(b)
	} else {
		b = countenvelope.NewRequest(id, sub.Filters).Marshal(b)
	}
	// log.T.F("{%s} sending %s", sub.Relay.URL, b)
	sub.live.Store(true)
	if err = <-sub.Relay.Write(b); chk.T(err) {
		sub.cancel()
		return errorf.E("failed to write: %w", err)
	}

	return nil
}
