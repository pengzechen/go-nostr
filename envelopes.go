package nostr

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/mailru/easyjson"
	jwriter "github.com/mailru/easyjson/jwriter"
	"github.com/tidwall/gjson"
)

func ParseMessage(message []byte) Envelope {
	firstComma := bytes.Index(message, []byte{','})
	if firstComma == -1 {
		return nil
	}
	label := message[0:firstComma]
	var v Envelope
	switch {
	case bytes.Contains(label, []byte("EVENT")):
		v = &EventEnvelope{}
	case bytes.Contains(label, []byte("REQ")):
		v = &ReqEnvelope{}
	case bytes.Contains(label, []byte("NOTICE")):
		x := NoticeEnvelope("")
		v = &x
	case bytes.Contains(label, []byte("EOSE")):
		x := EOSEEnvelope("")
		v = &x
	case bytes.Contains(label, []byte("OK")):
		v = &OKEnvelope{}
	case bytes.Contains(label, []byte("AUTH")):
		v = &AuthEnvelope{}
	case bytes.Contains(label, []byte("CLOSE")):
		x := CloseEnvelope("")
		v = &x
	default:
		return nil
	}

	if err := v.UnmarshalJSON(message); err != nil {
		return nil
	}
	return v
}

type Envelope interface {
	Label() string
	UnmarshalJSON([]byte) error
	MarshalJSON() ([]byte, error)
}

type EventEnvelope struct {
	SubscriptionID *string
	Event
}

var (
	_ Envelope = (*EventEnvelope)(nil)
	_ Envelope = (*ReqEnvelope)(nil)
	_ Envelope = (*NoticeEnvelope)(nil)
	_ Envelope = (*EOSEEnvelope)(nil)
	_ Envelope = (*CloseEnvelope)(nil)
	_ Envelope = (*OKEnvelope)(nil)
	_ Envelope = (*AuthEnvelope)(nil)
)

func (_ EventEnvelope) Label() string { return "EVENT" }

func (v *EventEnvelope) UnmarshalJSON(data []byte) error {
	r := gjson.ParseBytes(data)
	arr := r.Array()
	switch len(arr) {
	case 2:
		return easyjson.Unmarshal([]byte(arr[1].Raw), &v.Event)
	case 3:
		v.SubscriptionID = &arr[1].Str
		return easyjson.Unmarshal([]byte(arr[2].Raw), &v.Event)
	default:
		return fmt.Errorf("failed to decode EVENT envelope")
	}
}

func (v EventEnvelope) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	w.RawString(`["EVENT",`)
	if v.SubscriptionID != nil {
		w.RawString(`"` + *v.SubscriptionID + `",`)
	}
	v.MarshalEasyJSON(&w)
	w.RawString(`]`)
	return w.BuildBytes()
}

type ReqEnvelope struct {
	SubscriptionID string
	Filters
}

func (_ ReqEnvelope) Label() string { return "REQ" }

func (v *ReqEnvelope) UnmarshalJSON(data []byte) error {
	r := gjson.ParseBytes(data)
	arr := r.Array()
	if len(arr) < 3 {
		return fmt.Errorf("failed to decode REQ envelope: missing filters")
	}
	v.SubscriptionID = arr[1].Str
	v.Filters = make(Filters, len(arr)-2)
	f := 0
	for i := 2; i < len(arr); i++ {
		if err := easyjson.Unmarshal([]byte(arr[i].Raw), &v.Filters[f]); err != nil {
			return fmt.Errorf("%w -- on filter %d", err, f)
		}
		f++
	}

	return nil
}

func (v ReqEnvelope) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	w.RawString(`["REQ",`)
	w.RawString(`"` + v.SubscriptionID + `"`)
	for _, filter := range v.Filters {
		w.RawString(`,`)
		filter.MarshalEasyJSON(&w)
	}
	w.RawString(`]`)
	return w.BuildBytes()
}

type NoticeEnvelope string

func (_ NoticeEnvelope) Label() string { return "NOTICE" }

func (v *NoticeEnvelope) UnmarshalJSON(data []byte) error {
	r := gjson.ParseBytes(data)
	arr := r.Array()
	switch len(arr) {
	case 2:
		*v = NoticeEnvelope(arr[1].Str)
		return nil
	default:
		return fmt.Errorf("failed to decode NOTICE envelope")
	}
}

func (v NoticeEnvelope) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	w.RawString(`["NOTICE",`)
	w.Raw(json.Marshal(string(v)))
	w.RawString(`]`)
	return w.BuildBytes()
}

type EOSEEnvelope string

func (_ EOSEEnvelope) Label() string { return "EOSE" }

func (v *EOSEEnvelope) UnmarshalJSON(data []byte) error {
	r := gjson.ParseBytes(data)
	arr := r.Array()
	switch len(arr) {
	case 2:
		*v = EOSEEnvelope(arr[1].Str)
		return nil
	default:
		return fmt.Errorf("failed to decode EOSE envelope")
	}
}

func (v EOSEEnvelope) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	w.RawString(`["EOSE",`)
	w.Raw(json.Marshal(string(v)))
	w.RawString(`]`)
	return w.BuildBytes()
}

type CloseEnvelope string

func (_ CloseEnvelope) Label() string { return "CLOSE" }

func (v *CloseEnvelope) UnmarshalJSON(data []byte) error {
	r := gjson.ParseBytes(data)
	arr := r.Array()
	switch len(arr) {
	case 2:
		*v = CloseEnvelope(arr[1].Str)
		return nil
	default:
		return fmt.Errorf("failed to decode CLOSE envelope")
	}
}

func (v CloseEnvelope) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	w.RawString(`["CLOSE",`)
	w.Raw(json.Marshal(string(v)))
	w.RawString(`]`)
	return w.BuildBytes()
}

type OKEnvelope struct {
	EventID string
	OK      bool
	Reason  *string
}

func (_ OKEnvelope) Label() string { return "OK" }

func (v *OKEnvelope) UnmarshalJSON(data []byte) error {
	r := gjson.ParseBytes(data)
	arr := r.Array()
	if len(arr) < 3 {
		return fmt.Errorf("failed to decode OK envelope: missing fields")
	}
	v.EventID = arr[1].Str
	v.OK = arr[2].Raw == "true"

	if len(arr) > 3 {
		v.Reason = &arr[3].Str
	}

	return nil
}

func (v OKEnvelope) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	w.RawString(`["OK",`)
	w.RawString(`"` + v.EventID + `",`)
	ok := "false"
	if v.OK {
		ok = "true"
	}
	w.RawString(ok)
	if v.Reason != nil {
		w.RawString(`,`)
		w.Raw(json.Marshal(v.Reason))
	}
	w.RawString(`]`)
	return w.BuildBytes()
}

type AuthEnvelope struct {
	Challenge *string
	Event     Event
}

func (_ AuthEnvelope) Label() string { return "AUTH" }

func (v *AuthEnvelope) UnmarshalJSON(data []byte) error {
	r := gjson.ParseBytes(data)
	arr := r.Array()
	if len(arr) < 2 {
		return fmt.Errorf("failed to decode Auth envelope: missing fields")
	}
	if arr[1].IsObject() {
		return easyjson.Unmarshal([]byte(arr[1].Raw), &v.Event)
	} else {
		v.Challenge = &arr[1].Str
	}
	return nil
}

func (v AuthEnvelope) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	w.RawString(`["AUTH",`)
	if v.Challenge != nil {
		w.Raw(json.Marshal(*v.Challenge))
	} else {
		v.Event.MarshalEasyJSON(&w)
	}
	w.RawString(`]`)
	return w.BuildBytes()
}
