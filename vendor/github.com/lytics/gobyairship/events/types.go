package events

import "errors"

// WrongType is returned by per-Type methods on Event if method called doesn't
// match the Event's type.
var WrongType = errors.New("wrong type for event")

// Type of Event. Events contain per-Type methods to properly decode the Event
// body into a the desired Type.
type Type string

const (
	TypePush                   Type = "PUSH_BODY"
	TypeOpen                   Type = "OPEN"
	TypeSend                   Type = "SEND"
	TypeClose                  Type = "CLOSE"
	TypeTagChange              Type = "TAG_CHANGE"
	TypeUninstall              Type = "UNINSTALL"
	TypeFirst                  Type = "FIRST_OPEN"
	TypeCustom                 Type = "CUSTOM"
	TypeLocation               Type = "LOCATION"
	TypeRichDelivery           Type = "RICH_DELIVERY"
	TypeRichRead               Type = "RICH_READ"
	TypeRichDelete             Type = "RICH_DELETE"
	TypeInAppMessageDisplay    Type = "IN_APP_MESSAGE_DISPLAY"
	TypeInAppMessageResolution Type = "IN_APP_MESSAGE_RESOLUTION"
	TypeInAppMessageExpiration Type = "IN_APP_MESSAGE_EXPIRATION"
)

type Device struct {
	Amazon    string `json:"amazon_channel,omitempty"`
	Android   string `json:"android_channel,omitempty"`
	IOS       string `json:"ios_channel,omitempty"`
	NamedUser string `json:"named_user_id,omitempty"`
}
