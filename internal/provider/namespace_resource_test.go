package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccNamespaceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccNamespaceResourceConfig("test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("flink_appmanager_namespace.test", "name", "test"),
					resource.TestCheckResourceAttr("flink_appmanager_namespace.test", "state", "ACTIVE"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccNamespaceResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "flink_appmanager_namespace" "test" {
 provider = flink-appmanager
 name =  %[1]q
}
`, name)
}
