package events

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// LimitExceeded is returned when the number of simultaneous connections to
// Urban Airship's Event API is exceeded. The API responds with a 402 Payment
// Required status which is translated into this error.
var LimitExceeded = errors.New("request was rate limited")

// Event is the envelope for a single even from Urban Airship's event stream.
// Users should inspect the Event's Type and call the corresponding method to
// receive a typed event body.
type Event struct {
	// ID uniquely identifies the event.
	ID       string    `json:"id"`
	Type     Type      `json:"type"`
	Occurred time.Time `json:"occurred"`

	// Processed is when the event was ingested by Urban Airship. There may be
	// lag between when the event occurred, and when it was processed.
	Processed time.Time `json:"processed"`

	// Offset is the event's location in the stream. Used to resume the stream
	// after severing a connection. Clients should store this value for the case
	// that the connection is severed.
	Offset uint64 `json:"offset,string"`

	// Body is the raw event body. Use the Type specific methods to unmarshal the
	// body.
	Body   json.RawMessage `json:"body"`
	Device *Device         `json:"device,omitempty"`
}

type Push struct {
	// PushID is the unique identifier for the push, included in responses to the
	// push API.
	PushID string `json:"push_id"`

	// GroupID is an optional identifier of the group this push is associated
	// with; group IDs are created by both automation and push to local time.
	GroupID string `json:"group_id"`
}

type PushBody struct {
	Push

	// Payload is the specification of the push as sent via the API.
	Payload []byte `json:"payload"`
}

func (e *Event) PushBody() (*PushBody, error) {
	if e.Type != TypePush {
		return nil, WrongType
	}
	p := PushBody{}
	if err := json.Unmarshal(e.Body, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

type Open struct {
	// LastDelivered contains the push identifier of the last notification Urban
	// Airship attempted to deliver to this device, if known. It may also include
	// a group identifier if the push was scheduled to the device’s local time or
	// if the push was an automation rule.
	//
	LastDelivered *Push `json:"last_delivered,omitempty"`

	// Triggering is present if the event was associated with a push. An
	// object containing the push ID of that notification. It may also include a
	// group ID if the push was a push to local time or automation rule.
	//
	TriggeringPush *Push `json:"triggering_push,omitempty"`

	// SessionID is an identifier for the "session" of user activity. This key
	// will be absent if the application was initialized while backgrounded.
	SessionID string `json:"session_id"`
}

// Open returns an Open struct for OPEN events. Non-OPEN events will return
// the WrongType error.
func (e *Event) Open() (*Open, error) {
	if e.Type != TypeOpen {
		return nil, WrongType
	}
	o := Open{}
	if err := json.Unmarshal(e.Body, &o); err != nil {
		return nil, err
	}
	return &o, nil
}

// Send events are emitted for each device identified by the audience selection
// of a push. device will be present in the event to specify which channel
// received the push.
type Send struct {
	Push

	// VariantID is only present if the notification was sent as part of an
	// experiment.  Identifies the payload ultimately sent to a device.
	VariantID *int `json:"variant_id,omitempty"`
}

// Send returns a Send struct for SEND events. Non-SEND events will return the
// WrongType error.
func (e *Event) Send() (*Send, error) {
	if e.Type != TypeSend {
		return nil, WrongType
	}
	s := Send{}
	if err := json.Unmarshal(e.Body, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// Close events are emitted Each time a user closes the application. Note that
// close events are often latent, as they may not be delivered over the network
// until much later.
type Close struct {
	SessionID string `json:"session_id"`
}

func (e *Event) Close() (*Close, error) {
	if e.Type != TypeClose {
		return nil, WrongType
	}
	c := Close{}
	if err := json.Unmarshal(e.Body, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// TagChange events are emitted Each time a tag is added or removed from a
// channel.
type TagChange struct {
	// Add maps tag groups to tags. The set of tag group/tag pairs in this object
	// define the tags added to the device.
	Add map[string][]string `json:"add"`

	// Remove maps tag groups to tags. The set of tag group/tag pairs in this
	// object define the tags removed from the device.
	Remove map[string][]string `json:"remove"`

	// Current maps tag groups to tags. The set of tag group/tag pairs in this
	// object define the current state of the device AFTER the mutation has taken
	// effect.
	Current map[string][]string `json:"current"`
}

func (e *Event) TagChange() (*TagChange, error) {
	if e.Type != TypeTagChange {
		return nil, WrongType
	}
	t := TagChange{}
	if err := json.Unmarshal(e.Body, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// Location events include the latitude and longitude of the device.
type Location struct {
	Lat json.Number `json:"latitude"`
	Lon json.Number `json:"longitude"`

	// Foreground indicates whether the application was foregrounded when the
	// event fired.
	Foreground bool `json:"foreground"`
}

func (e *Event) Location() (*Location, error) {
	if e.Type != TypeLocation {
		return nil, WrongType
	}
	loc := Location{}
	if err := json.Unmarshal(e.Body, &loc); err != nil {
		return nil, err
	}
	return &loc, nil
}

type InAppMessageDisplay struct {
	Push

	// A triggering push is present if the user started the current session by opening
	// a push notification.
	TriggeringPush Push `json:"triggering_push"`
}

func (e *Event) InAppMessageDisplay() (*InAppMessageDisplay, error) {
	if e.Type != TypeInAppMessageDisplay {
		return nil, WrongType
	}
	disp := InAppMessageDisplay{}
	if err := json.Unmarshal(e.Body, &disp); err != nil {
		return nil, err
	}
	return &disp, nil
}

type InAppMessageResolution struct {
	InAppMessageDisplay

	TimeSent time.Time `json:"time_sent"`

	// Type indicates how the In-app message was resolved, and can take on one
	// of the following values: BUTTON_CLICK, MESSAGE_CLICK, TIMED_OUT, USER_DISMISSED
	Type string `json:"type"`

	// Duration is the amount of time for which the message was displayed, in milliseconds.
	Duration int64 `json:"duration"`

	// Additional optional fields describing the button that was clicked:
	ButtonID          string `json:"button_id,omitempty"`
	ButtonGroup       string `json:"button_group,omitempty"`
	ButtonDescription string `json:"button_description,omitempty"`
}

func (e *Event) InAppMessageResolution() (*InAppMessageResolution, error) {
	if e.Type != TypeInAppMessageResolution {
		return nil, WrongType
	}
	res := InAppMessageResolution{}
	if err := json.Unmarshal(e.Body, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

type InAppMessageExpiration struct {
	InAppMessageResolution

	// ReplacingPush is present if Type is equal to REPLACED. It identifies
	// the push specification defining the In-App message which should replace the
	// current message.
	ReplacingPush Push `json:"replacing_push,omitempty"`
}

func (e *Event) InAppMessageExpiration() (*InAppMessageExpiration, error) {
	if e.Type != TypeInAppMessageExpiration {
		return nil, WrongType
	}
	exp := InAppMessageExpiration{}
	if err := json.Unmarshal(e.Body, &exp); err != nil {
		return nil, err
	}
	return &exp, nil
}

func (e *Event) RichEvent() (*Push, error) {
	if e.Type != TypeRichDelete && e.Type != TypeRichDelivery && e.Type != TypeRichRead {
		return nil, WrongType
	}
	p := Push{}
	if err := json.Unmarshal(e.Body, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// Response streams Events from a Fetch call.
type Response struct {
	// ID is the UA-Operation-Id header from Urban Airship's response.
	ID string

	out  chan *Event
	body io.ReadCloser

	mu     *sync.Mutex
	closed chan struct{}
	err    error
}

// NewResponse creates an events iterator from an http.Response. Fetch is a
// shortcut for creating a Response, but users can manually create a Response
// from a custom HTTP request with this function.
func NewResponse(resp *http.Response) (*Response, error) {
	if resp.StatusCode == 402 {
		return nil, LimitExceeded
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected non-200 response: %d", resp.StatusCode)
	}
	r := &Response{
		ID:     resp.Header.Get("UA-Operation-Id"),
		out:    make(chan *Event, 10), // provide some buffering
		body:   resp.Body,
		mu:     new(sync.Mutex),
		closed: make(chan struct{}),
	}
	go func() {
		// Always close Event chan to indicate to callers that response is done.
		defer close(r.out)
		dec := json.NewDecoder(r.body)
		for {
			var ev Event
			if err := dec.Decode(&ev); err != nil {
				select {
				case <-r.closed:
					//TODO Only ignore "closed" errors
					return
				default:
					r.mu.Lock()
					defer r.mu.Unlock()
					r.err = err
					return
				}
			}
			select {
			case r.out <- &ev:
			case <-r.closed:
				return
			}
		}
	}()
	return r, nil
}

// Events returns a chan that emits Events until closed. Events is safe for
// concurrent calls and shares an underlying chan. This means events are not
// duplicated between multiple receivers.
func (r *Response) Events() <-chan *Event { return r.out }

// Close the events stream. Safe to call concurrently.
func (r *Response) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	select {
	case <-r.closed:
		return
	default:
		close(r.closed)
		r.body.Close()
	}
}

// Err returns the error which caused the event stream to end or nil. May be
// checked when the chan returned by Events() is closed. Safe for concurrent
// access.
func (r *Response) Err() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.err
}
