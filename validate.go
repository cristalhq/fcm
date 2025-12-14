package fcm

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var bareTopicNamePattern = regexp.MustCompile("^[a-zA-Z0-9-_.~%]+$")

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
