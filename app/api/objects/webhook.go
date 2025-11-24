package objects

/*
enabledboolean

Optional

Set this property to true to enable the Event Webhook or false to disable it.
urlstring
required

Set this property to the URL where you want the Event Webhook to send event data.
group_resubscribeboolean

Optional

Set this property to true to receive group resubscribe events. Group resubscribes occur when recipients resubscribe to a specific unsubscribe group by updating their subscription preferences. You must enable Subscription Tracking to receive this type of event.
deliveredboolean

Optional

Set this property to true to receive delivered events. Delivered events occur when a message has been successfully delivered to the receiving server.
group_unsubscribeboolean

Optional

Set this property to true to receive group unsubscribe events. Group unsubscribes occur when recipients unsubscribe from a specific unsubscribe group either by direct link or by updating their subscription preferences. You must enable Subscription Tracking to receive this type of event.
spam_reportboolean

Optional

Set this property to true to receive spam report events. Spam reports occur when recipients mark a message as spam.
bounceboolean

Optional

Set this property to true to receive bounce events. A bounce occurs when a receiving server could not or would not accept a message.
deferredboolean

Optional

Set this property to true to receive deferred events. Deferred events occur when a recipient's email server temporarily rejects a message.
unsubscribeboolean

Optional

Set this property to true to receive unsubscribe events. Unsubscribes occur when recipients click on a message's subscription management link. You must enable Subscription Tracking to receive this type of event.
processedboolean

Optional

Set this property to true to receive processed events. Processed events occur when a message has been received by Twilio SendGrid and the message is ready to be delivered.
openboolean

Optional

Set this property to true to receive open events. Open events occur when a recipient has opened the HTML message. You must enable Open Tracking to receive this type of event.
clickboolean

Optional

Set this property to true to receive click events. Click events occur when a recipient clicks on a link within the message. You must enable Click Tracking to receive this type of event.
droppedboolean

Optional

Set this property to true to receive dropped events. Dropped events occur when your message is not delivered by Twilio SendGrid. Dropped events are accomponied by a reason property, which indicates why the message was dropped. Reasons for a dropped message include: Invalid SMTPAPI header, Spam Content (if spam checker app enabled), Unsubscribed Address, Bounced Address, Spam Reporting Address, Invalid, Recipient List over Package Quota.
friendly_namestring or null

Optional

Optionally set this property to a friendly name for the Event Webhook. A friendly name may be assigned to each of your webhooks to help you differentiate them. The friendly name is for convenience only. You should use the webhook id property for any programmatic tasks.
oauth_client_idstring or null

Optional

Set this property to the OAuth client ID that SendGrid will pass to your OAuth server or service provider to generate an OAuth access token. When passing data in this property, you must also include the oauth_token_url property.
oauth_client_secretstring or null

Optional

Set this property to the OAuth client secret that SendGrid will pass to your OAuth server or service provider to generate an OAuth access token. This secret is needed only once to create an access token. SendGrid will store the secret, allowing you to update your client ID and Token URL without passing the secret to SendGrid again. When passing data in this field, you must also include the oauth_client_id and oauth_token_url properties.
oauth_token_urlstring or null

Optional

Set this property to the URL where SendGrid will send the OAuth client ID and client secret to generate an OAuth access token. This should be your OAuth server or service provider. When passing data in this field, you must also include the oauth_client_id property.

{
  "enabled": true,
  "url": "https://example.com/webhook-endpoint",
  "group_resubscribe": true,
  "delivered": false,
  "group_unsubscribe": true,
  "spam_report": true,
  "bounce": true,
  "deferred": true,
  "unsubscribe": true,
  "processed": false,
  "open": true,
  "click": true,
  "dropped": true,
  "friendly_name": "Engagement Webhook",
  "oauth_client_id": "a835e7210bbb47edbfa71bdfc909b2d7",
  "oauth_token_url": "https://oauthservice.example.com",
  "id": "77d4a5da-7015-11ed-a1eb-0242ac120002",
  "created_date": "2023-01-01T12:00:00Z",
  "updated_date": "2023-02-15T10:00:00Z"
}
*/

// Webhook represents a SendGrid-style webhook object.
type Webhook struct {
	Enabled             bool    `json:"enabled,omitempty"`
	URL                 string  `json:"url" validate:"required"`
	GroupResubscribe    bool    `json:"group_resubscribe,omitempty"`
	Delivered           bool    `json:"delivered,omitempty"`
	AccountStatusChange bool    `json:"account_status_change,omitempty"`
	GroupUnsubscribe    bool    `json:"group_unsubscribe,omitempty"`
	SpamReport          bool    `json:"spam_report,omitempty"`
	Bounce              bool    `json:"bounce,omitempty"`
	Deferred            bool    `json:"deferred,omitempty"`
	Unsubscribe         bool    `json:"unsubscribe,omitempty"`
	Processed           bool    `json:"processed,omitempty"`
	Open                bool    `json:"open,omitempty"`
	Click               bool    `json:"click,omitempty"`
	Dropped             bool    `json:"dropped,omitempty"`
	FriendlyName        *string `json:"friendly_name,omitempty"`
	ID                  *string `json:"id,omitempty"`
	OauthClientID       *string `json:"oauth_client_id,omitempty"`
	OauthClientSecret   *string `json:"oauth_client_secret,omitempty"`
	OauthTokenURL       *string `json:"oauth_token_url,omitempty"`
}
