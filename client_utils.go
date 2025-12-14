package fcm

import (
	"context"
	_ "embed"
	"errors"
	"maps"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func newHTTPClient(rawCreds []byte) (*http.Client, error) {
	trans, err := newTransport(rawCreds)
	if err != nil {
		return nil, err
	}
	return &http.Client{Transport: trans}, nil
}

type parameterTransport struct {
	userAgent     string
	quotaProject  string
	requestReason string

	base http.RoundTripper
}

func (t *parameterTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rt := t.base
	if rt == nil {
		return nil, errors.New("transport: no Transport specified")
	}

	newReq := *req
	newReq.Header = make(http.Header)
	maps.Copy(newReq.Header, req.Header)

	return rt.RoundTrip(&newReq)
}

func newTransport(rawCreds []byte) (http.RoundTripper, error) {
	paramTransport := &parameterTransport{
		base: http.DefaultTransport.(*http.Transport).Clone(),
	}
	var trans http.RoundTripper = paramTransport

	creds, err := internalCreds(rawCreds)
	if err != nil {
		return nil, err
	}

	trans = &oauth2.Transport{
		Base:   trans,
		Source: creds.TokenSource,
	}
	return trans, nil
}

func internalCreds(rawCreds []byte) (*google.Credentials, error) {
	return credentialsFromJSON(rawCreds)
}

// credentialsFromJSON returns a google.Credentials from the JSON data
//
// - A self-signed JWT flow will be executed if the following conditions are
// met:
//
//	(1) At least one of the following is true:
//	    (a) Scope for self-signed JWT flow is enabled
//	    (b) Audiences are explicitly provided by users
//	(2) No service account impersontation
//
// - Otherwise, executes standard OAuth 2.0 flow
// More details: google.aip.dev/auth/4111
func credentialsFromJSON(data []byte) (*google.Credentials, error) {
	ctx := context.Background()

	var params google.CredentialsParams
	params.Scopes = firebaseScopes

	oauth2Client := oauth2.NewClient(ctx, nil)
	params.TokenURL = google.Endpoint.TokenURL
	ctx = context.WithValue(ctx, oauth2.HTTPClient, oauth2Client)

	// By default, a standard OAuth 2.0 token source is created
	cred, err := google.CredentialsFromJSONWithParams(ctx, data, params)
	if err != nil {
		return nil, err
	}
	return cred, nil
}

var firebaseScopes = []string{
	"https://www.googleapis.com/auth/cloud-platform",
	"https://www.googleapis.com/auth/datastore",
	"https://www.googleapis.com/auth/devstorage.full_control",
	"https://www.googleapis.com/auth/firebase",
	"https://www.googleapis.com/auth/identitytoolkit",
	"https://www.googleapis.com/auth/userinfo.email",
}
