package events

import (
	"errors"
	"fmt"
	"net/http"
)

const DefaultEventsURL = "https://connect.urbanairship.com/api/events/"

var evurl = DefaultEventsURL

// SetURL allows overriding the default URL for Urban Airship's Event stream
// and returns the previous value. Passing an empty string will just return the
// current value without changing it.
func SetURL(url string) string {
	old := evurl
	if len(url) > 0 {
		evurl = url
	}
	return old
}

// Client used to fetch events. Usually *gobyairship.Client.
type Client interface {
	Post(url string, body interface{}, extra http.Header) (*http.Response, error)
}

// Start indicates whether to start at the earliest or latest offset. See
// Request for details.
type Start string

const (
	StartFirst Start = "EARLIEST"
	StartLast  Start = "LATEST"

	// Start from a specific offset (Offset must be specified in Fetch).
	StartOffset Start = ""
)

// DeviceType can be specified in a Filter to receive events for specific types
// of devices.
type DeviceType string

const (
	DeviceAmazon  DeviceType = "amazon"
	DeviceAndroid DeviceType = "android"
	DeviceIOS     DeviceType = "ios"
	DeviceUser    DeviceType = "named_user"
	deviceUnknown DeviceType = "unknown"
)

type Filter struct {
	Types        []Type       `json:"types,omitempty"`
	DeviceTypes  []DeviceType `json:"device_types,omitempty"`
	Notification []Push       `json:"notification,omitempty"`
	Devices      []Device     `json:"devices,omitempty"`
	Latency      int64        `json:"latency,omitempty"`
}

type SubsetType string

const (
	SubsetTypePartition SubsetType = "PARTITION"
	SubsetTypeSample    SubsetType = "SAMPLE"
)

type Subset struct {
	Type       SubsetType `json:"type"`
	Proportion *float64   `json:"proportion,omitempty"`
	Count      *int       `json:"count,omitempty"`
	Selection  *int       `json:"selection,omitempty"`
}

// SubsetPartition creates a Subset with "count" deterministic partitions of
// which this request will receive the "selection" partition.
func SubsetPartition(count, selection int) *Subset {
	return &Subset{Type: SubsetTypePartition, Count: &count, Selection: &selection}
}

// SubsetSample creates a random sampling Subset whose proportion should be
// between 0 and 1.
func SubsetSample(proportion float64) *Subset {
	return &Subset{Type: SubsetTypeSample, Proportion: &proportion}
}

// Validate returns an error if Subset is invalid otherwise nil.
func (s *Subset) Validate() error {
	if s == nil {
		// It's valid to not specify a subset
		return nil
	}
	switch s.Type {
	case SubsetTypePartition:
		if s.Count == nil || s.Selection == nil {
			return errors.New("count and selection must be set for partition subsets")
		}
		if *s.Count < 1 {
			return errors.New("count < 1")
		}
		if *s.Selection < 0 || *s.Selection >= *s.Count {
			return fmt.Errorf("selection must be [0,%d)", *s.Count)
		}
		if s.Proportion != nil {
			return errors.New("proportion must not be set for partition subsets")
		}
	case SubsetTypeSample:
		if s.Proportion == nil {
			return errors.New("proportion must be set for sample subsets")
		}
		if *s.Proportion < 0 || *s.Proportion > 1 {
			return fmt.Errorf("proportion %f not between [0,1]", *s.Proportion)
		}
		if s.Count != nil || s.Selection != nil {
			return errors.New("count and selection must not be set for sample subsets")
		}
	default:
		return fmt.Errorf("invalid subset type: %s", s.Type)
	}
	return nil
}

// Request is an Urban Airship Events API request. The Fetch function will
// create one internally, or you can manually create your own and submit it via
// the gobyairship.Client's Post method.
type Request struct {
	// Start is one of “EARLIEST” or “LATEST”. Specifies that the stream should
	// start at the beginning or the end of the application’s data window. Only
	// specify one of Offset and Start.
	Start Start `json:"start,omitempty"`

	// Offset specifies where to start streaming. Each Event specifies its offset
	// which can be used in subsequent requests to resume from where the previous
	// request ended.
	Offset *uint64 `json:"resume_offset,omitempty"`

	// Filters specifies the criteria an event must meet to be returned in the
	// response. Filters are unioned.
	Filters []*Filter `json:"filters,omitempty"`

	// Subset allows iterating over a subset of events based on either random
	// sampling or deterministic partitioning. See Subset type for details.
	Subset *Subset `json:"subset,omitempty"`
}

// Validate returns nil if the request is valid or an error if there's an
// issue.
func (r *Request) Validate() error {
	if r.Start != StartOffset && r.Offset != nil {
		return fmt.Errorf("only specify one of Start or Offset: start=%s offset=%d", r.Start, *r.Offset)
	}
	if r.Start != StartOffset && r.Start != StartFirst && r.Start != StartLast {
		return fmt.Errorf("start must be one of %q, %q, or %q", StartFirst, StartLast, StartOffset)
	}
	if err := r.Subset.Validate(); err != nil {
		return err
	}
	return nil
}

// Fetch events using a Client. Filters and subset may be nil to fetch all
// events. If error is non-nil Response will stream events until Close is
// called.
func Fetch(c Client, st Start, offset uint64, su *Subset, filters ...*Filter) (*Response, error) {
	req := &Request{Start: st, Subset: su, Filters: filters}
	if st == StartOffset {
		req.Offset = &offset
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Override Accept header with ndjson type
	extra := http.Header{"Accept": []string{"application/vnd.urbanairship+x-ndjson;version=3;"}}

	// Valid request, post to API
	resp, err := c.Post(evurl, req, extra)
	if err != nil {
		return nil, err
	}

	// Valid response, return events iterator
	return NewResponse(resp)
}
