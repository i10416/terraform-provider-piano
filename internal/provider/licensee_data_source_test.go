// Copyright (c) Yoichiro Ito <contact.110416@gmail.com>
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccLicenseeDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccExampleDataSourceConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.piano_licensee.test",
						tfjsonpath.New("aid"),
						knownvalue.StringExact("example"),
					),
					statecheck.ExpectKnownValue(
						"data.piano_licensee.test",
						tfjsonpath.New("licensee_id"),
						knownvalue.StringExact("example"),
					),
				},
			},
		},
	})
}

const testAccExampleDataSourceConfig = `
data "piano_licensee" "test" {
  aid = "example"
  licensee_id = "example"
}
`
