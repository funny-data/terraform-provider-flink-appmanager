terraform {
  required_providers {
    flink-appmanager = {
      source = "registry.terraform.io/funny-data/flink-appmanager"
    }
  }
}

provider "flink-appmanager" {
  endpoint      = "http://flink-appmanager"
  namespace = "test"
  wait_timeout = 180
}

resource "flink_appmanager_namespace" "test" {
  provider = flink-appmanager
  name = "test"
}

resource "flink_appmanager_deployment_target" "test" {
  provider = flink-appmanager
  depends_on = [
    flink_appmanager_namespace.test
  ]

  name = "test"
  namespace = "default"
}

resource "flink_appmanager_session_cluster" "test" {
  provider = flink-appmanager
  depends_on = [
    flink_appmanager_deployment_target.test
  ]

  name = "test"
  deployment_target_name = "test"
  flink_version = "1.14.4"
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