package fcm

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var (
	bareTopicNamePattern  = regexp.MustCompile("^[a-zA-Z0-9-_.~%]+$")
	colorPattern          = regexp.MustCompile("^#[0-9a-fA-F]{6}$")
	colorWithAlphaPattern = regexp.MustCompile("^#[0-9a-fA-F]{6}([0-9a-fA-F]{2})?$")
)

func validateMessage(message *Message) error {
	if message == nil {
		return errors.New("message must not be nil")
	}

	targets := countNonEmpty(message.Token, message.Condition, message.Topic)
	if targets != 1 {
		return errors.New("exactly one of token, topic or condition must be specified")
	}

	if message.Topic != "" {
		bt := strings.TrimPrefix(message.Topic, "/topics/")
		if !bareTopicNamePattern.MatchString(bt) {
			return errors.New("malformed topic name")
		}
	}

	if err := validateNotification(message.Notification); err != nil {
		return err
	}
	if err := validateAndroidConfig(message.Android); err != nil {
		return err
	}
	if err := validateWebpushConfig(message.Webpush); err != nil {
		return err
	}
	if err := validateAPNSConfig(message.APNS); err != nil {
		return err
	}
	return nil
}

func validateNotification(notification *Notification) error {
	if notification == nil {
		return nil
	}

	if image := notification.ImageURL; image != "" {
		if err := isValidURL(image); err != nil {
			return fmt.Errorf("invalid image URL: %q", image)
		}
	}
	return nil
}

func validateAndroidConfig(config *AndroidConfig) error {
	switch {
	case config == nil:
		return nil

	case config.TTL != nil && config.TTL.Seconds() < 0:
		return errors.New("ttl duration must not be negative")

	case config.Priority != "" && config.Priority != "normal" && config.Priority != "high":
		return errors.New("priority must be 'normal' or 'high'")

	default:
		return validateAndroidNotification(config.Notification)
	}
}

func validateAndroidNotification(notification *AndroidNotification) error {
	switch {
	case notification == nil:
		return nil

	case notification.Color != "" && !colorPattern.MatchString(notification.Color):
		return errors.New("color must be in the #RRGGBB form")

	case len(notification.TitleLocArgs) > 0 && notification.TitleLocKey == "":
		return errors.New("titleLocKey is required when specifying titleLocArgs")

	case len(notification.BodyLocArgs) > 0 && notification.BodyLocKey == "":
		return errors.New("bodyLocKey is required when specifying bodyLocArgs")
	}

	if image := notification.ImageURL; image != "" {
		if err := isValidURL(image); err != nil {
			return fmt.Errorf("invalid image URL: %q", image)
		}
	}

	for _, timing := range notification.VibrateTimingMillis {
		if timing < 0 {
			return errors.New("vibrateTimingMillis must not be negative")
		}
	}

	return validateLightSettings(notification.LightSettings)
}

func validateLightSettings(light *LightSettings) error {
	switch {
	case light == nil:
		return nil

	case !colorWithAlphaPattern.MatchString(light.Color):
		return errors.New("color must be in #RRGGBB or #RRGGBBAA form")

	case light.LightOnDurationMillis < 0:
		return errors.New("lightOnDuration must not be negative")

	case light.LightOffDurationMillis < 0:
		return errors.New("lightOffDuration must not be negative")

	default:
		return nil
	}
}

func validateAPNSConfig(config *APNSConfig) error {
	if config == nil {
		return nil
	}

	if config.FCMOptions != nil {
		image := config.FCMOptions.ImageURL
		if image != "" {
			if err := isValidURL(image); err != nil {
				return fmt.Errorf("invalid image URL: %q", image)
			}
		}
	}
	return validateAPNSPayload(config.Payload)
}

func validateAPNSPayload(payload *APNSPayload) error {
	if payload == nil {
		return nil
	}

	m := payload.standardFields()
	for k := range payload.CustomData {
		if _, contains := m[k]; contains {
			return fmt.Errorf("multiple specifications for the key %q", k)
		}
	}
	return validateAps(payload.Aps)
}

func validateAps(aps *Aps) error {
	if aps == nil {
		return nil
	}
	if aps.Alert != nil && aps.AlertString != "" {
		return errors.New("multiple alert specifications")
	}

	if aps.CriticalSound != nil {
		if aps.Sound != "" {
			return errors.New("multiple sound specifications")
		}
		if aps.CriticalSound.Volume < 0 || aps.CriticalSound.Volume > 1 {
			return errors.New("critical sound volume must be in the interval [0, 1]")
		}
	}

	m := aps.standardFields()
	for k := range aps.CustomData {
		if _, contains := m[k]; contains {
			return fmt.Errorf("multiple specifications for the key: %q", k)
		}
	}
	return validateApsAlert(aps.Alert)
}

func validateApsAlert(alert *ApsAlert) error {
	switch {
	case alert == nil:
		return nil

	case len(alert.TitleLocArgs) > 0 && alert.TitleLocKey == "":
		return errors.New("titleLocKey is required when specifying titleLocArgs")

	case len(alert.SubTitleLocArgs) > 0 && alert.SubTitleLocKey == "":
		return errors.New("subtitleLocKey is required when specifying subtitleLocArgs")

	case len(alert.LocArgs) > 0 && alert.LocKey == "":
		return errors.New("locKey is required when specifying locArgs")

	default:
		return nil
	}
}

func validateWebpushConfig(webpush *WebpushConfig) error {
	if webpush == nil || webpush.Notification == nil {
		return nil
	}

	dir := webpush.Notification.Direction
	if dir != "" && dir != "ltr" && dir != "rtl" && dir != "auto" {
		return errors.New("direction must be 'ltr', 'rtl' or 'auto'")
	}

	m := webpush.Notification.standardFields()
	for k := range webpush.Notification.CustomData {
		if _, contains := m[k]; contains {
			return fmt.Errorf("multiple specifications for the key %q", k)
		}
	}

	if webpush.FCMOptions != nil {
		link := webpush.FCMOptions.Link
		p, err := url.ParseRequestURI(link)
		if err != nil {
			return fmt.Errorf("invalid link URL: %q", link)
		} else if p.Scheme != "https" {
			return fmt.Errorf("invalid link URL: %q; want scheme: %q", link, "https")
		}
	}
	return nil
}

func isValidURL(link string) error {
	_, err := url.ParseRequestURI(link)
	return err
}

func countNonEmpty(ss ...string) int {
	count := 0
	for _, s := range ss {
		if s != "" {
			count++
		}
	}
	return count
}
