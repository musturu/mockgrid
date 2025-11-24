package objects

/*


asm_group_id	X	X	X	X	X
bounce_classification	X
attempt			X
category	X	X	X	X	X
email	X	X	X	X	X
event	X	X	X	X	X
ip		X
marketing_campaign_id	X	X	X	X	X
marketing_campaign_name	X	X	X	X	X
pool					X
reason	X		X	X
response		X
sg_event_id	X	X	X	X	X
sg_message_id	X*	X	X	X	X
smtp-id	X	X	X	X	X
status	X
timestamp	X	X	X	X	X
tls	X	X
unique_args	X	X	X	X	X
*/

type DelieryEvent struct {
	ASM_Group_ID            string            `json:"asm_group_id,omitempty"`
	Bounce_Classification   string            `json:"bounce_classification,omitempty"`
	Attempt                 int               `json:"attempt,omitempty"`
	Category                []string          `json:"category,omitempty"`
	Email                   string            `json:"email,omitempty"`
	Event                   string            `json:"event,omitempty"`
	IP                      string            `json:"ip,omitempty"`
	Marketing_Campaign_ID   string            `json:"marketing_campaign_id,omitempty"`
	Marketing_Campaign_Name string            `json:"marketing_campaign_name,omitempty"`
	Pool                    string            `json:"pool,omitempty"`
	Reason                  string            `json:"reason,omitempty"`
	Response                string            `json:"response,omitempty"`
	Sg_Event_ID             string            `json:"sg_event_id,omitempty"`
	Sg_Message_ID           string            `json:"sg_message_id,omitempty"`
	Smtp_ID                 string            `json:"smtp-id,omitempty"`
	Status                  string            `json:"status,omitempty"`
	Timestamp               int64             `json:"timestamp,omitempty"`
	TLS                     int               `json:"tls,omitempty"`
	Unique_Args             map[string]string `json:"unique_args,omitempty"`
}
