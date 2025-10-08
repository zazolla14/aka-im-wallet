package constant

import "github.com/1nterdigital/grpc-protocol/constant"

const (
	MountConfigFilePath = "CONFIG_PATH"
	DeploymentType      = "DEPLOYMENT_TYPE"
	KUBERNETES          = "kubernetes"
	ETCD                = "etcd"
)

const (
	NormalUser = 1
	AdminUser  = 2
)

const (
	RpcOpUserID   = constant.OpUserID
	RpcOpUserType = "opUserType"
)

const RpcCustomHeader = constant.RpcCustomHeader

type ContextKey string

const (
	ContextKeyVersion ContextKey = "version"
)
