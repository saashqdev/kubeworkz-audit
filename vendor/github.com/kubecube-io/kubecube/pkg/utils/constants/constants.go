/*
Copyright 2024 KubeWorkz Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package constants

const (
	// KubeWorkz all the begin
	KubeWorkz = "kubeworkz"

	// Warden is willing to kubeworkz
	Warden = "warden"

	// ApiPathRoot the root api route
	ApiPathRoot = "/api/v1/kube"

	// CubeNamespace kubeworkz default namespace
	CubeNamespace = "kubeworkz-system"

	// PivotCluster pivot cluster name
	// Deprecated
	PivotCluster = "pivot-cluster"

	// LocalCluster the internal cluster where program stand with
	LocalCluster = "_local_cluster"

	// DefaultPivotCubeClusterIPSvc default pivot kube svc
	DefaultPivotCubeClusterIPSvc = "kubeworkz.kubeworkz-system:7443"

	DefaultAuditURL = "http://audit.kubeworkz-system:8888/api/v1/kube/audit/kube"
)

// http content
const (
	HttpHeaderContentType        = "Content-type"
	HttpHeaderContentDisposition = "Content-Disposition"
	HttpHeaderContentTypeOctet   = "application/octet-stream"

	ImpersonateUserKey  = "Impersonate-User"
	ImpersonateGroupKey = "Impersonate-Group"
)

// audit and user constant
const (
	EventName          = "event"
	EventTypeUserWrite = "userwrite"
	EventResourceType  = "resourceType"
	EventAccountId     = "accountId"
	EventObjectName    = "objectName"
	EventRespBody      = "responseBody"

	AuthorizationHeader        = "Authorization"
	DefaultTokenExpireDuration = 3600 // 1 hour
)

// k8s api resources
const (
	K8sResourceVersion   = "v1"
	K8sResourceNamespace = "namespaces"
	K8sResourcePod       = "pods"

	K8sKindClusterRole    = "ClusterRole"
	K8sKindRole           = "Role"
	K8sKindServiceAccount = "ServiceAccount"

	K8sGroupRBAC = "rbac.authorization.k8s.io"
)

// rbac related constant
const (
	PlatformAdmin = "platform-admin"
	TenantAdmin   = "tenant-admin"
	ProjectAdmin  = "project-admin"
	Reviewer      = "reviewer"

	TenantAdminCluster  = "tenant-admin-cluster"
	ProjectAdminCluster = "project-admin-cluster"
	ReviewerCluster     = "reviewer-cluster"

	PlatformAdminAgLabel = "rbac.authorization.k8s.io/aggregate-to-platform-admin"
	TenantAdminAgLabel   = "rbac.authorization.k8s.io/aggregate-to-tenant-admin"
	ProjectAdminAgLabel  = "rbac.authorization.k8s.io/aggregate-to-project-admin"
	ReviewerAgLabel      = "rbac.authorization.k8s.io/aggregate-to-reviewer"
)

const (
	// ClusterLabel indicates the resource which cluster relate with
	ClusterLabel = "kubeworkz.io/cluster"

	// TenantLabel represent which tenant resource relate with
	TenantLabel = "kubeworkz.io/tenant"

	// ProjectLabel represent which project resource relate with
	ProjectLabel = "kubeworkz.io/project"

	// CubeQuotaLabel point to CubeResourceQuota
	CubeQuotaLabel = "kubeworkz.io/quota"

	// RbacLabel indicates the resource of rbac is related with kubeworkz
	RbacLabel = "kubeworkz.io/rbac"
	// RoleLabel indicates the role of rbac policy
	RoleLabel = "kubeworkz.io/role"

	// CrdLabel indicates the crds kubeworkz need to dispatch
	CrdLabel = "kubeworkz.io/crds"

	// SyncAnnotation use for sync logic of warden
	SyncAnnotation = "kubeworkz.io/sync"
)

const (
	// CubeNodeTaint is node taint that managed by KubeWorkz
	CubeNodeTaint = "node.kubeworkz.io"
)

// hnc related conest
const (
	// HncInherited means resource is inherited form upon namespace by hnc
	HncInherited = "hnc.x-k8s.io/inherited-from"
)

// rbac role verbs
const (
	// AllVerb all verbs
	AllVerb = "*"
	// CreateVerb create resource
	CreateVerb = "create"
	// DeleteVerb delete resource
	DeleteVerb = "delete"
	// ListVerb list resource
	ListVerb = "list"
)
