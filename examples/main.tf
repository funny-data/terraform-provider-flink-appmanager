terraform {
  required_providers {
    fam = {
      source = "registry.terraform.io/funny-data/flink-appmanager"
    }
  }
}

provider "fam" {
  endpoint     = "http://flink-appmanager"
  wait_timeout = 180
}

resource "flink_appmanager_namespace" "test" {
  provider = fam
  name     = "test"
}

resource "flink_appmanager_deployment_target" "test" {
  provider = fam
  depends_on = [
    flink_appmanager_namespace.test
  ]

  name      = "test"
  namespace = "test"
}

resource "flink_appmanager_session_cluster" "test" {
  provider = fam
  depends_on = [
    flink_appmanager_deployment_target.test
  ]

  name                    = "test"
  namespace               = "test"
  deployment_target_name  = "test"
  flink_image_tag         = "1.14.4-scala_2.12-java11-1"
  number_of_task_managers = 1
  flink_configuration = {
    "high-availability" : "flink-kubernetes"
    "execution.checkpointing.externalized-checkpoint-retention" = "RETAIN_ON_CANCELLATION"
    "execution.checkpointing.interval"                          = "60s"
    "execution.checkpointing.min-pause"                         = "60s"
    "state.backend"                                             = "true"
  }
  resources = {
    taskmanager = {
      cpu    = 1
      memory = "1G"
    }
    jobmanager = {
      cpu    = 1
      memory = "1G"
    }
  }
}