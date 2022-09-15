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
			// Create and Read DeploymentTargetResourceModel Resource
			{
				Config: testAccDeploymentTargetResourceConfig("test", "default"),
				Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr("flink_appmanager_deployment_target.test", "name", "test"),
					resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr("flink_appmanager_deployment_target.test", "namespace", "test")),
					resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr("flink_appmanager_deployment_target.test", "k8s_namespace", "default")),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccDeploymentTargetResourceConfig(name string, k8snamespace string) string {
	return fmt.Sprintf(`
resource "flink_appmanager_namespace" "test" {
 provider = fam

 name =  "test"
}

resource "flink_appmanager_deployment_target" "test" {
provider = fam
depends_on = [
 flink_appmanager_namespace.test
]

name = %[1]q
namespace = flink_appmanager_namespace.test.name
k8s_namespace = %[2]q
}
`, name, k8snamespace)
}
