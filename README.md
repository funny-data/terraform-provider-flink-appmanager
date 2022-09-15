# Flink-AppManager terraform Provider

插件开发文档: https://www.terraform.io/plugin/framework

## 前提
环境
- go
- terraform

## 本地打包编译
创建`~/.terraformrc`文件,注意将`<user>`替换成本机名称
```text
provider_installation {

  dev_overrides {
      "registry.terraform.io/funny-data/flink-appmanager"" = "/Users/<user>/go/bin"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

在项目目录下执行命令
```shell
go mod tidy
go install
```

## 开发测试

进入项目`example`目录下,需在`main.tf`中配置`FlinkAppManager`的主机地址`host`

FlinkAppManager Provider参数配置说明
- `host`: FlinkAppManager主机地址,参数示例: `http://flink-appmanager`
- `wait_timeout`: 资源操作超时时间,默认180秒,参数示例: `180`
- `wait_interval`: 资源操作检查间隔,默认3秒,参数示例: `3`

执行以下命令进行AppManager的资源管理
```shell
# 创建资源
terraform apply

# 销毁资源
terraform destroy
```

状态导入说明
```shell
# 导入test的namespace
terraform import flink_appmanager_namespace.test test
 
# 导入test的部署目标
terraform import flink_appmanager_deployment_target.test test

# 导入test的集群
terraform import flink_appmanager_session_cluster.test test
```

## 自动化测试

这里需要将端点指定为测试平台
```shell
export FLINK_APPMANAGER_ENDPOINT=http://flink-appmanager

make testacc
```