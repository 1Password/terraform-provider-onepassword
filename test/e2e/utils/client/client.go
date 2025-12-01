package client

import (
	"context"
	"os"

	"github.com/1Password/terraform-provider-onepassword/v2/internal/onepassword"
)

func CreateTestClient(ctx context.Context) (onepassword.Client, error) {
	return onepassword.NewClient(ctx, onepassword.ClientConfig{
		ConnectHost:         os.Getenv("OP_CONNECT_HOST"),
		ConnectToken:        os.Getenv("OP_CONNECT_TOKEN"),
		ServiceAccountToken: os.Getenv("OP_SERVICE_ACCOUNT_TOKEN"),
		Account:             os.Getenv("OP_ACCOUNT"),
		OpCLIPath:           "op",
		ProviderUserAgent:   "terraform-provider-onepassword/test",
	})
}
