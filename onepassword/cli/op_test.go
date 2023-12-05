package cli

import (
	"errors"
	"testing"

	"github.com/1Password/connect-sdk-go/onepassword"
)

func TestWithRetry(t *testing.T) {
	op := &OP{}
	tests := map[string]struct {
		fakeAction func() (*onepassword.Item, error)
		validate   func(item *onepassword.Item, err error)
	}{
		"should fail when error other than 409": {
			fakeAction: func() (*onepassword.Item, error) {
				return nil, errors.New("failed to perform action")
			},
			validate: func(item *onepassword.Item, err error) {
				if err == nil {
					t.Error("Action should fail when error is other than 409")
				}
				if item != nil {
					t.Error("Item should be nil when error is other than 409")
				}
			},
		},
		"should fail when error is 409": {
			fakeAction: func() (*onepassword.Item, error) {
				return nil, errors.New("409 Conflict error")
			},
			validate: func(item *onepassword.Item, err error) {
				if err == nil {
					t.Error("Action should fail it when error is 409")
				}
				if item != nil {
					t.Error("Item should be nil when error is 409")
				}
			},
		},
		"should succeed": {
			fakeAction: func() (*onepassword.Item, error) {
				return &onepassword.Item{}, nil
			},
			validate: func(item *onepassword.Item, err error) {
				if err != nil {
					t.Errorf("Action should succeed, but got an error: %s", err.Error())
				}
				if item == nil {
					t.Error("Item should not be nil")
				}
			},
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			test.validate(op.withRetry(test.fakeAction))
		})
	}
}
