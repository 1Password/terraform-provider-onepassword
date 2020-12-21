package onepassword

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	// Provider name for single configuration testing
	ProviderName = "onepassword"
)

var testAccProviders map[string]*schema.Provider
var testAccProviderFactories map[string]func() (*schema.Provider, error)
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()

	testAccProviders = map[string]*schema.Provider{
		ProviderName: testAccProvider,
	}

	testAccProviderFactories = map[string]func() (*schema.Provider, error){
		ProviderName: func() (*schema.Provider, error) { return Provider(), nil }, //nolint:unparam
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider()
}

func TestProvider_HasResources(t *testing.T) {
	expectedResources := []string{
		"onepassword_item",
	}

	resources := Provider().ResourcesMap
	if len(expectedResources) != len(resources) {
		t.Errorf("There are an unexpected number of registered resources. Expected %v got %v", len(expectedResources), len(resources))
	}

	for _, resource := range expectedResources {
		if _, ok := resources[resource]; !ok {
			t.Errorf("An expected resource was not registered")
		}
		if resources[resource] == nil {
			t.Errorf("A resource cannot have a nil schema")
		}
	}
}

func TestProvider_HasDataSources(t *testing.T) {
	expectedDataSources := []string{
		"onepassword_item",
	}

	dataSources := Provider().DataSourcesMap
	if len(expectedDataSources) != len(dataSources) {
		t.Errorf("There are an unexpected number of registered data sources. Expected %v got %v", len(expectedDataSources), len(dataSources))
	}

	for _, resource := range expectedDataSources {
		if _, ok := dataSources[resource]; !ok {
			t.Errorf("An expected data source was not registered")
		}
		if dataSources[resource] == nil {
			t.Errorf("A data source cannot have a nil schema")
		}
	}
}
