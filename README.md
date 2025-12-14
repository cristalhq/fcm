# fcm

[![build-img]][build-url]
[![pkg-img]][pkg-url]
[![reportcard-img]][reportcard-url]
[![coverage-img]][coverage-url]
[![version-img]][version-url]

Firebase Cloud Messaging client in Go. A lightweight alternative to [firebase.google.com/go/v4](https://pkg.go.dev/firebase.google.com/go/v4).

## Rationale

`firebase.google.com/go/v4` has tremendous dependencies, resulting in nearly 1,000,000 lines of code (as of v4.18.0).
There are many features in the original library, but sending a push notification via FCM requires just a single HTTP request.
That’s exactly what this library provides.

## Features

* Simple API.
* Clean and tested code.
* Dependency-free (only [golang.org/x/oauth2](golang.org/x/oauth2))

## Install

Go version 1.24+

```
go get github.com/cristalhq/fcm
```

## Example

Build new token:

```go
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
	panic(err)
}

fmt.Printf("pushID: %s\n", pushID)
```

Also see examples: [example_test.go](https://github.com/cristalhq/fcm/blob/main/example_test.go).

## Documentation

See [these docs][pkg-url].

## License

[MIT License](LICENSE).

[build-img]: https://github.com/cristalhq/fcm/workflows/build/badge.svg
[build-url]: https://github.com/cristalhq/fcm/actions
[pkg-img]: https://pkg.go.dev/badge/cristalhq/fcm
[pkg-url]: https://pkg.go.dev/github.com/cristalhq/fcm
[reportcard-img]: https://goreportcard.com/badge/cristalhq/fcm
[reportcard-url]: https://goreportcard.com/report/cristalhq/fcm
[coverage-img]: https://codecov.io/gh/cristalhq/fcm/branch/main/graph/badge.svg
[coverage-url]: https://codecov.io/gh/cristalhq/fcm
[version-img]: https://img.shields.io/github/v/release/cristalhq/fcm
[version-url]: https://github.com/cristalhq/fcm/releases
