package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	sdk "github.com/1password/onepassword-sdk-go"
)

type VaultAPIMock struct {
	mock.Mock
}

func (v *VaultAPIMock) List(ctx context.Context) ([]sdk.VaultOverview, error) {
	args := v.Called(ctx)
	return args.Get(0).([]sdk.VaultOverview), args.Error(1)
}
