package util

const (
	// SystemNSsPrefix has preifx for system namespaces
	// we are assuming all the system namespaces start with `kube-`
	SystemNSsPrefix = "kube-"
	// IncludeKubeSysNSAnn specifies if the configs should be created
	// in the system namespaces as well
	IncludeKubeSysNSAnn = "krius.dev/includesystemns"
)
