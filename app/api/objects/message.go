package objects

/*
from_emailstring

Optional
msg_idstring

Optional
subjectstring

Optional
to_emailstring

Optional
statusenum<string>

Optional
Possible values:
processed
delivered
not_delivered
opens_countinteger

Optional
clicks_countinteger

Optional
last_event_timestring

Optional

iso 8601 format
*/
// Message represents a SendGrid-style message object.
type Message struct {
	From          string `json:"from_email" validate:"required"`
	MsgID         string `json:"msg_id,omitempty"`
	Subject       string `json:"subject,omitempty"`
	To            string `json:"to_email,omitempty"`
	Status        string `json:"status,omitempty"`
	OpensCount    int    `json:"opens_count,omitempty"`
	ClicksCount   int    `json:"clicks_count,omitempty"`
	LastEventTime string `json:"last_event_time,omitempty"`
}
