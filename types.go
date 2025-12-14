package fcm

import (
	"encoding/json"
	"strings"
)

// Message to be sent via Firebase Cloud Messaging (FCM).
// A Message must specify exactly one of Token, Topic or Condition fields.
//
// See https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages for more details.
type Message struct {
	Data         map[string]string `json:"data,omitempty"`
	Notification *Notification     `json:"notification,omitempty"`

	Token     string `json:"token,omitempty"`
	Topic     string `json:"-"`
	Condition string `json:"condition,omitempty"`
}

func (m Message) IsValid() error {
	return validateMessage(&m)
}

func (m *Message) MarshalJSON() ([]byte, error) {
	type messageWrapper Message

	tmp := &struct {
		BareTopic string `json:"topic,omitempty"`
		*messageWrapper
	}{
		BareTopic:      strings.TrimPrefix(m.Topic, "/topics/"),
		messageWrapper: (*messageWrapper)(m),
	}
	return json.Marshal(tmp)
}

func (m *Message) UnmarshalJSON(b []byte) error {
	type messageWrapper Message

	tmp := struct {
		BareTopic string `json:"topic,omitempty"`
		*messageWrapper
	}{
		messageWrapper: (*messageWrapper)(m),
	}
	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}
	m.Topic = tmp.BareTopic

	return nil
}

// Notification is the basic notification template to use across all platforms.
//
// See https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages#notification for more details.
type Notification struct {
	Title    string `json:"title,omitempty"`
	Body     string `json:"body,omitempty"`
	ImageURL string `json:"image,omitempty"`
}
