package relay

import (
	"net/http"

	"orly.dev/pkg/encoders/event"
	"orly.dev/pkg/utils/context"
)

// AcceptEvent determines whether an incoming event should be accepted for
// processing based on authentication requirements.
//
// # Parameters
//
//   - c: the context of the request
//
//   - ev: pointer to the event structure
//
//   - hr: HTTP request related to the event (if any)
//
//   - authedPubkey: public key of the authenticated user (if any)
//
//   - remote: remote address from where the event was received
//
// # Return Values
//
//   - accept: boolean indicating whether the event should be accepted
//
//   - notice: string providing a message or error notice
//
//   - afterSave: function to execute after saving the event (if applicable)
//
// # Expected Behaviour:
//
// - If authentication is required and no public key is provided, reject the
// event.
//
// - Otherwise, accept the event for processing.
func (s *Server) AcceptEvent(
	c context.T, ev *event.E, hr *http.Request, authedPubkey []byte,
	remote string,
) (accept bool, notice string, afterSave func()) {
	// if auth is required and the user is not authed, reject
	if s.AuthRequired() && len(authedPubkey) == 0 {
		return
	}
	accept = true
	return
}
