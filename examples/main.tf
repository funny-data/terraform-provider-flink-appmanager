terraform {
  required_providers {
    appmanager = {
      source = "xmfunny.com/funnydb/appmanager"
    }
  }
}

provider "appmanager" {
  endpoint      = "http://flink-appmanager"
  namespace = "test"
  wait_timeout = 180
}

resource "am_namespace" "test" {
  provider = appmanager
  name = "test"
}

resource "am_deployment_target" "test" {
  provider = appmanager
  depends_on = [
    am_namespace.test
  ]

  name = "test"
  namespace = "default"
}

resource "am_session_cluster" "test" {
  provider = appmanager
  depends_on = [
    am_deployment_target.test
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