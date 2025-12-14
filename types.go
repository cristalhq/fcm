package fcm

import (
	"encoding/json"
	"fmt"
	"maps"
	"strconv"
	"strings"
	"time"
)

const rfc3339Zulu = "2006-01-02T15:04:05.000000000Z"

// Message to be sent via Firebase Cloud Messaging (FCM).
// A Message must specify exactly one of Token, Topic or Condition fields.
//
// See https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages
type Message struct {
	Data         map[string]string `json:"data,omitempty"`
	Notification *Notification     `json:"notification,omitempty"`
	Android      *AndroidConfig    `json:"android,omitempty"`
	Webpush      *WebpushConfig    `json:"webpush,omitempty"`
	APNS         *APNSConfig       `json:"apns,omitempty"`
	FCMOptions   *FCMOptions       `json:"fcm_options,omitempty"`

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
// See https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages#notification
type Notification struct {
	Title    string `json:"title,omitempty"`
	Body     string `json:"body,omitempty"`
	ImageURL string `json:"image,omitempty"`
}

// AndroidConfig contains messaging options specific to the Android platform.
//
// See https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages#androidconfig
type AndroidConfig struct {
	CollapseKey           string               `json:"collapse_key,omitempty"`
	Priority              string               `json:"priority,omitempty"` // one of "normal" or "high"
	TTL                   *time.Duration       `json:"-"`
	RestrictedPackageName string               `json:"restricted_package_name,omitempty"`
	Data                  map[string]string    `json:"data,omitempty"` // if set, overrides [Message.Data] field.
	Notification          *AndroidNotification `json:"notification,omitempty"`
	FCMOptions            *AndroidFCMOptions   `json:"fcm_options,omitempty"`
	DirectBootOK          bool                 `json:"direct_boot_ok,omitempty"`
}

func (a *AndroidConfig) MarshalJSON() ([]byte, error) {
	var ttl string
	if a.TTL != nil {
		ttl = durationToString(*a.TTL)
	}

	type androidWrapper AndroidConfig

	tmp := &struct {
		TTL string `json:"ttl,omitempty"`
		*androidWrapper
	}{
		TTL:            ttl,
		androidWrapper: (*androidWrapper)(a),
	}
	return json.Marshal(tmp)
}

func (a *AndroidConfig) UnmarshalJSON(b []byte) error {
	type androidWrapper AndroidConfig

	tmp := struct {
		TTL string `json:"ttl,omitempty"`
		*androidWrapper
	}{
		androidWrapper: (*androidWrapper)(a),
	}
	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}
	if tmp.TTL != "" {
		ttl, err := stringToDuration(tmp.TTL)
		if err != nil {
			return err
		}
		a.TTL = &ttl
	}
	return nil
}

// AndroidNotification is a notification to send to Android devices.
//
// See https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages#androidnotification
type AndroidNotification struct {
	Title                 string                        `json:"title,omitempty"` // if set, overrides [Notification.Title] field.
	Body                  string                        `json:"body,omitempty"`  // if set, overrides [Notification.Body] field.
	Icon                  string                        `json:"icon,omitempty"`
	Color                 string                        `json:"color,omitempty"` // #RRGGBB format
	Sound                 string                        `json:"sound,omitempty"`
	Tag                   string                        `json:"tag,omitempty"`
	ClickAction           string                        `json:"click_action,omitempty"`
	BodyLocKey            string                        `json:"body_loc_key,omitempty"`
	BodyLocArgs           []string                      `json:"body_loc_args,omitempty"`
	TitleLocKey           string                        `json:"title_loc_key,omitempty"`
	TitleLocArgs          []string                      `json:"title_loc_args,omitempty"`
	ChannelID             string                        `json:"channel_id,omitempty"`
	Ticker                string                        `json:"ticker,omitempty"`
	Sticky                bool                          `json:"sticky,omitempty"`
	EventTimestamp        *time.Time                    `json:"-"`
	LocalOnly             bool                          `json:"local_only,omitempty"`
	Priority              AndroidNotificationPriority   `json:"-"`
	DefaultSound          bool                          `json:"default_sound,omitempty"`
	DefaultVibrateTimings bool                          `json:"default_vibrate_timings,omitempty"`
	DefaultLightSettings  bool                          `json:"default_light_settings,omitempty"`
	VibrateTimingMillis   []int64                       `json:"-"`
	Visibility            AndroidNotificationVisibility `json:"-"`
	NotificationCount     *int                          `json:"notification_count,omitempty"`
	LightSettings         *LightSettings                `json:"light_settings,omitempty"`
	ImageURL              string                        `json:"image,omitempty"`
	Proxy                 AndroidNotificationProxy      `json:"-"`
}

func (a *AndroidNotification) MarshalJSON() ([]byte, error) {
	var priority string
	if a.Priority != priorityUnknown {
		priorities := map[AndroidNotificationPriority]string{
			PriorityMin:     "PRIORITY_MIN",
			PriorityLow:     "PRIORITY_LOW",
			PriorityDefault: "PRIORITY_DEFAULT",
			PriorityHigh:    "PRIORITY_HIGH",
			PriorityMax:     "PRIORITY_MAX",
		}
		priority = priorities[a.Priority]
	}

	var visibility string
	if a.Visibility != visibilityUnknown {
		visibilities := map[AndroidNotificationVisibility]string{
			VisibilityPrivate: "PRIVATE",
			VisibilityPublic:  "PUBLIC",
			VisibilitySecret:  "SECRET",
		}
		visibility = visibilities[a.Visibility]
	}

	var proxy string
	if a.Proxy != proxyUnknown {
		proxies := map[AndroidNotificationProxy]string{
			ProxyAllow:             "ALLOW",
			ProxyDeny:              "DENY",
			ProxyIfPriorityLowered: "IF_PRIORITY_LOWERED",
		}
		proxy = proxies[a.Proxy]
	}

	var timestamp string
	if a.EventTimestamp != nil {
		timestamp = a.EventTimestamp.UTC().Format(rfc3339Zulu)
	}

	vibTimings := make([]string, 0, len(a.VibrateTimingMillis))
	for _, t := range a.VibrateTimingMillis {
		vibTimings = append(vibTimings, durationToString(time.Duration(t)*time.Millisecond))
	}

	type androidWrapper AndroidNotification
	tmp := &struct {
		EventTimestamp string   `json:"event_time,omitempty"`
		Priority       string   `json:"notification_priority,omitempty"`
		Visibility     string   `json:"visibility,omitempty"`
		Proxy          string   `json:"proxy,omitempty"`
		VibrateTimings []string `json:"vibrate_timings,omitempty"`
		*androidWrapper
	}{
		EventTimestamp: timestamp,
		Priority:       priority,
		Visibility:     visibility,
		Proxy:          proxy,
		VibrateTimings: vibTimings,
		androidWrapper: (*androidWrapper)(a),
	}
	return json.Marshal(tmp)
}

func (a *AndroidNotification) UnmarshalJSON(b []byte) error {
	type androidWrapper AndroidNotification
	tmp := struct {
		EventTimestamp string   `json:"event_time,omitempty"`
		Priority       string   `json:"notification_priority,omitempty"`
		Visibility     string   `json:"visibility,omitempty"`
		Proxy          string   `json:"proxy,omitempty"`
		VibrateTimings []string `json:"vibrate_timings,omitempty"`
		*androidWrapper
	}{
		androidWrapper: (*androidWrapper)(a),
	}
	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}

	if tmp.Priority != "" {
		priorities := map[string]AndroidNotificationPriority{
			"PRIORITY_MIN":     PriorityMin,
			"PRIORITY_LOW":     PriorityLow,
			"PRIORITY_DEFAULT": PriorityDefault,
			"PRIORITY_HIGH":    PriorityHigh,
			"PRIORITY_MAX":     PriorityMax,
		}
		if prio, ok := priorities[tmp.Priority]; ok {
			a.Priority = prio
		} else {
			return fmt.Errorf("unknown priority value: %q", tmp.Priority)
		}
	}

	if tmp.Visibility != "" {
		visibilities := map[string]AndroidNotificationVisibility{
			"PRIVATE": VisibilityPrivate,
			"PUBLIC":  VisibilityPublic,
			"SECRET":  VisibilitySecret,
		}
		if vis, ok := visibilities[tmp.Visibility]; ok {
			a.Visibility = vis
		} else {
			return fmt.Errorf("unknown visibility value: %q", tmp.Visibility)
		}
	}

	if tmp.Proxy != "" {
		proxies := map[string]AndroidNotificationProxy{
			"ALLOW":               ProxyAllow,
			"DENY":                ProxyDeny,
			"IF_PRIORITY_LOWERED": ProxyIfPriorityLowered,
		}
		if prox, ok := proxies[tmp.Proxy]; ok {
			a.Proxy = prox
		} else {
			return fmt.Errorf("unknown proxy value: %q", tmp.Proxy)
		}
	}

	if tmp.EventTimestamp != "" {
		ts, err := time.Parse(rfc3339Zulu, tmp.EventTimestamp)
		if err != nil {
			return err
		}

		a.EventTimestamp = &ts
	}

	vibTimings := make([]int64, 0, len(tmp.VibrateTimings))
	for _, t := range tmp.VibrateTimings {
		vibTime, err := stringToDuration(t)
		if err != nil {
			return err
		}

		millis := int64(vibTime / time.Millisecond)
		vibTimings = append(vibTimings, millis)
	}
	a.VibrateTimingMillis = vibTimings
	return nil
}

// AndroidNotificationPriority represents the priority levels of a notification.
type AndroidNotificationPriority int

const (
	priorityUnknown AndroidNotificationPriority = 0

	// PriorityMin is the lowest notification priority.
	// Notifications with this priority might not be shown to the user except under special circumstances, such as detailed notification logs.
	PriorityMin AndroidNotificationPriority = 1

	// PriorityLow is a lower notification priority.
	// The UI may choose to show the notifications smaller, or at a different position in the list, compared with notifications with PriorityDefault.
	PriorityLow AndroidNotificationPriority = 2

	// PriorityDefault is the default notification priority.
	// If the application does not prioritize its own notifications, use this value for all notifications.
	PriorityDefault AndroidNotificationPriority = 3

	// PriorityHigh is a higher notification priority.
	// Use this for more important notifications or alerts.
	// The UI may choose to show these notifications larger, or at a different position in the notification lists, compared with notifications with PriorityDefault.
	PriorityHigh AndroidNotificationPriority = 4

	// PriorityMax is the highest notification priority.
	// Use this for the application's most important items that require the user's prompt attention or input.
	PriorityMax AndroidNotificationPriority = 5
)

// AndroidNotificationVisibility represents the different visibility levels of a notification.
type AndroidNotificationVisibility int

const (
	visibilityUnknown AndroidNotificationVisibility = 0

	// VisibilityPrivate shows this notification on all lockscreens, but conceal sensitive or private information on secure lockscreens.
	VisibilityPrivate AndroidNotificationVisibility = 1

	// VisibilityPublic shows this notification in its entirety on all lockscreens.
	VisibilityPublic AndroidNotificationVisibility = 2

	// VisibilitySecret does not reveal any part of this notification on a secure lockscreen.
	VisibilitySecret AndroidNotificationVisibility = 3
)

// AndroidNotificationProxy to control when a notification may be proxied.
type AndroidNotificationProxy int

const (
	proxyUnknown AndroidNotificationProxy = 0

	// ProxyAllow tries to proxy this notification.
	ProxyAllow AndroidNotificationProxy = 1

	// ProxyDeny does not proxy this notification.
	ProxyDeny AndroidNotificationProxy = 2

	// ProxyIfPriorityLowered only tries to proxy this notification if its AndroidConfig's Priority was lowered from high to normal on the device.
	ProxyIfPriorityLowered AndroidNotificationProxy = 3
)

// LightSettings to control notification LED.
type LightSettings struct {
	Color                  string
	LightOnDurationMillis  int64
	LightOffDurationMillis int64
}

func (l *LightSettings) MarshalJSON() ([]byte, error) {
	clr, err := newColor(l.Color)
	if err != nil {
		return nil, err
	}

	tmp := struct {
		Color            *color `json:"color"`
		LightOnDuration  string `json:"light_on_duration"`
		LightOffDuration string `json:"light_off_duration"`
	}{
		Color:            clr,
		LightOnDuration:  durationToString(time.Duration(l.LightOnDurationMillis) * time.Millisecond),
		LightOffDuration: durationToString(time.Duration(l.LightOffDurationMillis) * time.Millisecond),
	}
	return json.Marshal(tmp)
}

func (l *LightSettings) UnmarshalJSON(b []byte) error {
	tmp := struct {
		Color            *color `json:"color"`
		LightOnDuration  string `json:"light_on_duration"`
		LightOffDuration string `json:"light_off_duration"`
	}{}
	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}

	on, err := stringToDuration(tmp.LightOnDuration)
	if err != nil {
		return err
	}

	off, err := stringToDuration(tmp.LightOffDuration)
	if err != nil {
		return err
	}

	l.Color = tmp.Color.toString()
	l.LightOnDurationMillis = int64(on / time.Millisecond)
	l.LightOffDurationMillis = int64(off / time.Millisecond)
	return nil
}

type color struct {
	Red   float64 `json:"red"`
	Green float64 `json:"green"`
	Blue  float64 `json:"blue"`
	Alpha float64 `json:"alpha"`
}

func newColor(clr string) (*color, error) {
	red, err := strconv.ParseInt(clr[1:3], 16, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", clr, err)
	}

	green, err := strconv.ParseInt(clr[3:5], 16, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", clr, err)
	}

	blue, err := strconv.ParseInt(clr[5:7], 16, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", clr, err)
	}

	alpha := int64(255)
	if len(clr) == 9 {
		alpha, err = strconv.ParseInt(clr[7:9], 16, 32)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", clr, err)
		}
	}

	return &color{
		Red:   float64(red) / 255.0,
		Green: float64(green) / 255.0,
		Blue:  float64(blue) / 255.0,
		Alpha: float64(alpha) / 255.0,
	}, nil
}

func (c *color) toString() string {
	red := int(c.Red * 255.0)
	green := int(c.Green * 255.0)
	blue := int(c.Blue * 255.0)
	alpha := int(c.Alpha * 255.0)
	if alpha == 255 {
		return fmt.Sprintf("#%X%X%X", red, green, blue)
	}
	return fmt.Sprintf("#%X%X%X%X", red, green, blue, alpha)
}

// AndroidFCMOptions contains additional options for features provided by the FCM Android SDK.
//
// See https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages#androidfcmoptions
type AndroidFCMOptions struct {
	AnalyticsLabel string `json:"analytics_label,omitempty"`
}

// WebpushConfig contains messaging options specific to the WebPush protocol.
// https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages#webpushconfig
//
// See https://tools.ietf.org/html/rfc8030#section-5
type WebpushConfig struct {
	Headers      map[string]string    `json:"headers,omitempty"`
	Data         map[string]string    `json:"data,omitempty"`
	Notification *WebpushNotification `json:"notification,omitempty"`
	FCMOptions   *WebpushFCMOptions   `json:"fcm_options,omitempty"`
}

// WebpushNotificationAction represents an action that can be performed upon receiving a WebPush notification.
type WebpushNotificationAction struct {
	Action string `json:"action,omitempty"`
	Title  string `json:"title,omitempty"`
	Icon   string `json:"icon,omitempty"`
}

// WebpushNotification is a notification to send via WebPush protocol.
//
// See https://developer.mozilla.org/en-US/docs/Web/API/notification/Notification
type WebpushNotification struct {
	Actions            []*WebpushNotificationAction `json:"actions,omitempty"`
	Title              string                       `json:"title,omitempty"` // if set, overrides [Notification.Title] field.
	Body               string                       `json:"body,omitempty"`  // if set, overrides [Notification.Body] field.
	Icon               string                       `json:"icon,omitempty"`
	Badge              string                       `json:"badge,omitempty"`
	Direction          string                       `json:"dir,omitempty"` // one of 'ltr' or 'rtl'
	Data               any                          `json:"data,omitempty"`
	Image              string                       `json:"image,omitempty"`
	Language           string                       `json:"lang,omitempty"`
	Renotify           bool                         `json:"renotify,omitempty"`
	RequireInteraction bool                         `json:"requireInteraction,omitempty"`
	Silent             bool                         `json:"silent,omitempty"`
	Tag                string                       `json:"tag,omitempty"`
	TimestampMillis    *int64                       `json:"timestamp,omitempty"`
	Vibrate            []int                        `json:"vibrate,omitempty"`
	CustomData         map[string]any
}

// standardFields creates a map containing all the fields except the custom data.
func (n *WebpushNotification) standardFields() map[string]any {
	m := make(map[string]any)
	addNonEmpty := func(key, value string) {
		if value != "" {
			m[key] = value
		}
	}
	addTrue := func(key string, value bool) {
		if value {
			m[key] = value
		}
	}
	if len(n.Actions) > 0 {
		m["actions"] = n.Actions
	}
	addNonEmpty("title", n.Title)
	addNonEmpty("body", n.Body)
	addNonEmpty("icon", n.Icon)
	addNonEmpty("badge", n.Badge)
	addNonEmpty("dir", n.Direction)
	addNonEmpty("image", n.Image)
	addNonEmpty("lang", n.Language)
	addTrue("renotify", n.Renotify)
	addTrue("requireInteraction", n.RequireInteraction)
	addTrue("silent", n.Silent)
	addNonEmpty("tag", n.Tag)
	if n.Data != nil {
		m["data"] = n.Data
	}
	if n.TimestampMillis != nil {
		m["timestamp"] = *n.TimestampMillis
	}
	if len(n.Vibrate) > 0 {
		m["vibrate"] = n.Vibrate
	}
	return m
}

func (n *WebpushNotification) MarshalJSON() ([]byte, error) {
	m := n.standardFields()
	for k, v := range n.CustomData {
		m[k] = v
	}
	return json.Marshal(m)
}

func (n *WebpushNotification) UnmarshalJSON(b []byte) error {
	type webpushNotificationWrapper WebpushNotification

	tmp := (*webpushNotificationWrapper)(n)
	if err := json.Unmarshal(b, tmp); err != nil {
		return err
	}
	allFields := make(map[string]any)
	if err := json.Unmarshal(b, &allFields); err != nil {
		return err
	}
	for k := range n.standardFields() {
		delete(allFields, k)
	}
	if len(allFields) > 0 {
		n.CustomData = allFields
	}
	return nil
}

// WebpushFCMOptions contains additional options for features provided by the FCM web SDK.
//
// See https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages#webpushfcmoptions
type WebpushFCMOptions struct {
	Link string `json:"link,omitempty"`
}

// APNSConfig contains messaging options specific to the Apple Push Notification Service (APNS).
// https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages#apnsconfig
//
// See https://developer.apple.com/library/content/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/CommunicatingwithAPNs.html
type APNSConfig struct {
	Headers           map[string]string `json:"headers,omitempty"`
	Payload           *APNSPayload      `json:"payload,omitempty"`
	FCMOptions        *APNSFCMOptions   `json:"fcm_options,omitempty"`
	LiveActivityToken string            `json:"live_activity_token,omitempty"`
}

// APNSPayload is the payload that can be included in an APNS message.
//
// The payload mainly consists of the aps dictionary. Additionally it may contain arbitrary
// key-values pairs as custom data fields.
//
// See https://developer.apple.com/library/content/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/PayloadKeyReference.html
type APNSPayload struct {
	Aps        *Aps           `json:"aps,omitempty"`
	CustomData map[string]any `json:"-"`
}

// standardFields creates a map containing all the fields except the custom data.
func (p *APNSPayload) standardFields() map[string]any {
	return map[string]any{"aps": p.Aps}
}

func (p *APNSPayload) MarshalJSON() ([]byte, error) {
	m := p.standardFields()
	maps.Copy(m, p.CustomData)
	return json.Marshal(m)
}

func (p *APNSPayload) UnmarshalJSON(b []byte) error {
	type apnsPayloadWrapper APNSPayload

	tmp := (*apnsPayloadWrapper)(p)
	if err := json.Unmarshal(b, tmp); err != nil {
		return err
	}
	allFields := make(map[string]any)
	if err := json.Unmarshal(b, &allFields); err != nil {
		return err
	}
	for k := range p.standardFields() {
		delete(allFields, k)
	}
	if len(allFields) > 0 {
		p.CustomData = allFields
	}
	return nil
}

// Aps represents the aps dictionary that may be included in an APNSPayload.
//
// Alert may be specified as a string (via the AlertString field), or as a struct (via the Alert field).
type Aps struct {
	AlertString      string         `json:"-"`
	Alert            *ApsAlert      `json:"-"`
	Badge            *int           `json:"badge,omitempty"`
	Sound            string         `json:"-"`
	CriticalSound    *CriticalSound `json:"-"`
	ContentAvailable bool           `json:"-"`
	MutableContent   bool           `json:"-"`
	Category         string         `json:"category,omitempty"`
	ThreadID         string         `json:"thread-id,omitempty"`
	CustomData       map[string]any `json:"-"`
}

// standardFields creates a map containing all the fields except the custom data.
func (a *Aps) standardFields() map[string]any {
	m := make(map[string]any)
	if a.Alert != nil {
		m["alert"] = a.Alert
	} else if a.AlertString != "" {
		m["alert"] = a.AlertString
	}
	if a.ContentAvailable {
		m["content-available"] = 1
	}
	if a.MutableContent {
		m["mutable-content"] = 1
	}
	if a.Badge != nil {
		m["badge"] = *a.Badge
	}
	if a.CriticalSound != nil {
		m["sound"] = a.CriticalSound
	} else if a.Sound != "" {
		m["sound"] = a.Sound
	}
	if a.Category != "" {
		m["category"] = a.Category
	}
	if a.ThreadID != "" {
		m["thread-id"] = a.ThreadID
	}
	return m
}

func (a *Aps) MarshalJSON() ([]byte, error) {
	m := a.standardFields()
	maps.Copy(m, a.CustomData)
	return json.Marshal(m)
}

func (a *Aps) UnmarshalJSON(b []byte) error {
	type apsWrapper Aps
	tmp := struct {
		AlertObject         *json.RawMessage `json:"alert,omitempty"`
		SoundObject         *json.RawMessage `json:"sound,omitempty"`
		ContentAvailableInt int              `json:"content-available,omitempty"`
		MutableContentInt   int              `json:"mutable-content,omitempty"`
		*apsWrapper
	}{
		apsWrapper: (*apsWrapper)(a),
	}
	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}
	a.ContentAvailable = (tmp.ContentAvailableInt == 1)
	a.MutableContent = (tmp.MutableContentInt == 1)
	if tmp.AlertObject != nil {
		if err := json.Unmarshal(*tmp.AlertObject, &a.Alert); err != nil {
			a.Alert = nil
			if err := json.Unmarshal(*tmp.AlertObject, &a.AlertString); err != nil {
				return fmt.Errorf("failed to unmarshal alert as a struct or a string: %w", err)
			}
		}
	}
	if tmp.SoundObject != nil {
		if err := json.Unmarshal(*tmp.SoundObject, &a.CriticalSound); err != nil {
			a.CriticalSound = nil
			if err := json.Unmarshal(*tmp.SoundObject, &a.Sound); err != nil {
				return fmt.Errorf("failed to unmarshal sound as a struct or a string")
			}
		}
	}

	allFields := make(map[string]any)
	if err := json.Unmarshal(b, &allFields); err != nil {
		return err
	}
	for k := range a.standardFields() {
		delete(allFields, k)
	}
	if len(allFields) > 0 {
		a.CustomData = allFields
	}
	return nil
}

// CriticalSound is the sound payload that can be included in an Aps.
type CriticalSound struct {
	Critical bool    `json:"-"`
	Name     string  `json:"name,omitempty"`
	Volume   float64 `json:"volume,omitempty"`
}

func (cs *CriticalSound) MarshalJSON() ([]byte, error) {
	type criticalSoundWrapper CriticalSound
	tmp := struct {
		CriticalInt int `json:"critical,omitempty"`
		*criticalSoundWrapper
	}{
		criticalSoundWrapper: (*criticalSoundWrapper)(cs),
	}
	if cs.Critical {
		tmp.CriticalInt = 1
	}
	return json.Marshal(tmp)
}

func (cs *CriticalSound) UnmarshalJSON(b []byte) error {
	type criticalSoundWrapper CriticalSound
	tmp := struct {
		CriticalInt int `json:"critical,omitempty"`
		*criticalSoundWrapper
	}{
		criticalSoundWrapper: (*criticalSoundWrapper)(cs),
	}
	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}
	cs.Critical = (tmp.CriticalInt == 1)
	return nil
}

// ApsAlert is the alert payload that can be included in an Aps.
//
// See https://developer.apple.com/library/content/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/PayloadKeyReference.html
type ApsAlert struct {
	Title           string   `json:"title,omitempty"` // if set, overrides [Notification.Title] field.
	SubTitle        string   `json:"subtitle,omitempty"`
	Body            string   `json:"body,omitempty"` // if set, overrides [Notification.Body] field.
	LocKey          string   `json:"loc-key,omitempty"`
	LocArgs         []string `json:"loc-args,omitempty"`
	TitleLocKey     string   `json:"title-loc-key,omitempty"`
	TitleLocArgs    []string `json:"title-loc-args,omitempty"`
	SubTitleLocKey  string   `json:"subtitle-loc-key,omitempty"`
	SubTitleLocArgs []string `json:"subtitle-loc-args,omitempty"`
	ActionLocKey    string   `json:"action-loc-key,omitempty"`
	LaunchImage     string   `json:"launch-image,omitempty"`
}

// APNSFCMOptions contains additional options for features provided by the FCM Aps SDK.
//
// See https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages#apnsfcmoptions
type APNSFCMOptions struct {
	AnalyticsLabel string `json:"analytics_label,omitempty"`
	ImageURL       string `json:"image,omitempty"`
}

// FCMOptions contains additional options to use across all platforms.
//
// See https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages#fcmoptions
type FCMOptions struct {
	AnalyticsLabel string `json:"analytics_label,omitempty"`
}

func durationToString(ms time.Duration) string {
	seconds := int64(ms / time.Second)
	nanos := int64((ms - time.Duration(seconds)*time.Second) / time.Nanosecond)
	if nanos > 0 {
		return fmt.Sprintf("%d.%09ds", seconds, nanos)
	}
	return fmt.Sprintf("%ds", seconds)
}

func stringToDuration(s string) (time.Duration, error) {
	segments := strings.Split(strings.TrimSuffix(s, "s"), ".")
	if len(segments) != 1 && len(segments) != 2 {
		return 0, fmt.Errorf("incorrect number of segments in ttl: %q", s)
	}

	seconds, err := strconv.ParseInt(segments[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s: %w", s, err)
	}

	ttl := time.Duration(seconds) * time.Second
	if len(segments) == 2 {
		nanos, err := strconv.ParseInt(strings.TrimLeft(segments[1], "0"), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("failed to parse %s: %w", s, err)
		}
		ttl += time.Duration(nanos) * time.Nanosecond
	}

	return ttl, nil
}
