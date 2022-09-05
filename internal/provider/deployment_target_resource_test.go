package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccDeploymentTargetResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read DeploymentTarget Resource
			{
				Config: testAccDeploymentTargetResourceConfig("test", "default"),
				Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr("flink_appmanager_deployment_target.test", "name", "test"),
					resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr("flink_appmanager_deployment_target.test", "namespace", "default")),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccDeploymentTargetResourceConfig(name string, ns string) string {
	return fmt.Sprintf(`
resource "flink_appmanager_namespace" "test" {
 provider = flink-appmanager

 name =  "test"
}

resource "flink_appmanager_deployment_target" "test" {
provider = appmanager
depends_on = [
 flink_appmanager_namespace.test
]

name = %[1]q
namespace = %[2]q
}
`, name, ns)
}
