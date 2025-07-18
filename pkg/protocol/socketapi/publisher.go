package socketapi

import (
	"orly.dev/pkg/encoders/envelopes/eventenvelope"
	"orly.dev/pkg/encoders/event"
	"orly.dev/pkg/encoders/filters"
	"orly.dev/pkg/interfaces/publisher"
	"orly.dev/pkg/protocol/ws"
	"orly.dev/pkg/utils/chk"
	"orly.dev/pkg/utils/log"
	"regexp"
	"sync"
)

const Type = "socketapi"

var (
	NIP20prefixmatcher = regexp.MustCompile(`^\w+: `)
)

// Map is a map of filters associated with a collection of ws.Listener
// connections.
type Map map[*ws.Listener]map[string]*filters.T

type W struct {
	*ws.Listener

	// If Cancel is true, this is a close command.
	Cancel bool

	// Id is the subscription Id. If Cancel is true, cancel the named
	// subscription, otherwise, cancel the publisher for the socket.
	Id string

	// The Receiver holds the event channel for receiving notifications or data
	// relevant to this WebSocket connection.
	Receiver event.C

	// Filters holds a collection of filters used to match or process events
	// associated with this WebSocket connection. It is used to determine which
	// notifications or data should be received by the subscriber.
	Filters *filters.T
}

func (w *W) Type() (typeName string) { return Type }

type Close struct {
	*ws.Listener
	Id string
}

// S is a structure that manages subscriptions and associated filters for
// websocket listeners. It uses a mutex to synchronize access to a map storing
// subscriber connections and their filter configurations.
type S struct {
	// Mx is the mutex for the Map.
	Mx sync.Mutex
	// Map is the map of subscribers and subscriptions from the websocket api.
	Map
}

var _ publisher.I = &S{}

func New() (publisher *S) { return &S{Map: make(Map)} }

func (p *S) Type() (typeName string) { return Type }

// Receive handles incoming messages to manage websocket listener subscriptions
// and associated filters.
//
// # Parameters
//
// - msg (publisher.Message): The incoming message to process; expected to be of
// type *W to trigger subscription management actions.
//
// # Expected behaviour
//
// - Checks if the message is of type *W.
//
// - If Cancel is true, removes a subscriber by ID or the entire listener.
//
// - Otherwise, adds the subscription to the map under a mutex lock.
//
// - Logs actions related to subscription creation or removal.
func (p *S) Receive(msg publisher.Message) {
	if m, ok := msg.(*W); ok {
		if m.Cancel {
			if m.Id == "" {
				p.removeSubscriber(m.Listener)
				log.T.F("removed listener %s", m.Listener.RealRemote())
			} else {
				p.removeSubscriberId(m.Listener, m.Id)
				log.T.F(
					"removed subscription %s for %s", m.Id,
					m.Listener.RealRemote(),
				)
			}
			return
		}
		p.Mx.Lock()
		if subs, ok := p.Map[m.Listener]; !ok {
			subs = make(map[string]*filters.T)
			subs[m.Id] = m.Filters
			p.Map[m.Listener] = subs
			log.T.F(
				"created new subscription for %s, %s", m.Listener.RealRemote(),
				m.Filters.Marshal(nil),
			)
		} else {
			subs[m.Id] = m.Filters
			log.T.F(
				"added subscription %s for %s", m.Id, m.Listener.RealRemote(),
			)
		}
		p.Mx.Unlock()
	}
}

// Deliver sends an event to all subscribers whose filters match the event
//
// # Parameters
//
// - ev (*event.E): The event to deliver to matching subscribers
//
// # Expected behaviour
//
// # Locks the mutex to synchronize access to subscriber data
//
// # Iterates over all websocket connections and their associated subscriptions
//
// # Checks if each subscription's filter matches the event being delivered
//
// # Creates an event envelope result for matching subscriptions
//
// # Writes the result to the corresponding websocket connection
//
// Logs details about event delivery and any errors encountered
func (p *S) Deliver(ev *event.E) {
	log.T.F("delivering event %0x to subscribers", ev.Id)
	var err error
	p.Mx.Lock()
	for w, subs := range p.Map {
		log.I.F("%v %s", subs, w.RealRemote())
		for id, subscriber := range subs {
			log.T.F(
				"subscriber %s\n%s", w.RealRemote(),
				subscriber.Marshal(nil),
			)
			if !subscriber.Match(ev) {
				continue
			}
			var res *eventenvelope.Result
			if res, err = eventenvelope.NewResultWith(id, ev); chk.E(err) {
				continue
			}
			if err = res.Write(w); chk.E(err) {
				continue
			}
			log.T.F("dispatched event %0x to subscription %s", ev.Id, id)
		}
	}
	p.Mx.Unlock()
}

// removeSubscriberId removes a specific subscription from a subscriber
// websocket.
func (p *S) removeSubscriberId(ws *ws.Listener, id string) {
	p.Mx.Lock()
	var subs map[string]*filters.T
	var ok bool
	if subs, ok = p.Map[ws]; ok {
		delete(p.Map[ws], id)
		_ = subs
		if len(subs) == 0 {
			delete(p.Map, ws)
		}
	}
	p.Mx.Unlock()
}

// removeSubscriber removes a websocket from the S collection.
func (p *S) removeSubscriber(ws *ws.Listener) {
	p.Mx.Lock()
	clear(p.Map[ws])
	delete(p.Map, ws)
	p.Mx.Unlock()
}
