package scanver

import (
	"context"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"golang.org/x/xerrors"
)

// Client is GitHub API client wrapper for scanver.
type Client struct {
	*github.Client
}

// LookupAccessToken looks up GitHub access token from environment variable.
//
// This function expects GitHub access token is stored in environment variable named `GITHUB_ACCESS_TOKEN`.
func LookupAccessToken() (string, error) {
	token, ok := os.LookupEnv("GITHUB_ACCESS_TOKEN")

	if !ok {
		return "", xerrors.New("Access token is missing. Please set GITHUB_ACCESS_TOKEN environment variable.")
	}

	return token, nil
}

// NewClient returns a new client from GitHub access token.
func NewClient(ctx context.Context, token string) *Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	return &Client{client}
}
