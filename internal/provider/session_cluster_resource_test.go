package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccSessionClusterResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read SessionClusterResourceModel Resource
			{
				Config: testAccSessionClusterResourceConfig("test", "test"),
				Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr("flink_appmanager_session_cluster.test", "name", "test"),
					resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr("flink_appmanager_session_cluster.test", "deployment_target_name", "test")),
					resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr("flink_appmanager_session_cluster.test", "state", "RUNNING")),
				),
			},
		},
	})
}

func testAccSessionClusterResourceConfig(name string, deploymentTargetName string) string {
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

 name = %[2]q
 namespace = flink_appmanager_namespace.test.name
 k8s_namespace = "default"
}

resource "flink_appmanager_session_cluster" "test" {
  provider = fam
  depends_on = [
    flink_appmanager_deployment_target.test
  ]

  name = %[1]q
  namespace = flink_appmanager_namespace.test.name
  deployment_target_name = flink_appmanager_deployment_target.test.name
  flink_image_tag = "1.14.4-scala_2.12-java11-1"
  number_of_task_managers = 1
  flink_configuration = {
    "high-availability": "flink-kubernetes"
    "execution.checkpointing.externalized-checkpoint-retention" = "RETAIN_ON_CANCELLATION"
    "execution.checkpointing.interval" = "60s"
    "execution.checkpointing.min-pause" = "60s"
    "state.backend" = "true"
  }
  resources = {
    taskmanager = {
      cpu = 1
      memory = "1G"
    }
    jobmanager = {
      cpu = 1
      memory = "1G"
    }
  }
}
`, name, deploymentTargetName)
}
