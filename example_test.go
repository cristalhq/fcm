package fcm_test

import (
	"context"
	_ "embed"

	"github.com/cristalhq/fcm"
)

func Example() {
	ctx := context.Background()

	creds := []byte("...") // your JSON file from Firebase project settings

	cfg := fcm.Config{
		ProjectID:   "example-android-app",
		Credentials: creds,
	}

	client, err := fcm.NewClient(cfg)
	if err != nil {
		panic(err)
	}

	deviceToken := "..."

	msg := &fcm.Message{
		Data: map[string]string{
			"force_show": "1",
		},
		Notification: &fcm.Notification{
			Title: "Test",
			Body:  "Push from https://github.com/cristalhq/fcm",
		},
		Token: deviceToken,
	}

	pushID, err := client.Send(ctx, msg)
	if err != nil {
		// skipping error handling for example test
	}

	_ = pushID // notification ID

	// Output:
}
