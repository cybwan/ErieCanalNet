package k8s

import "fmt"

var (
	errListingNamespaces = fmt.Errorf("failed to list monitored namespaces")
	errServiceNotFound   = fmt.Errorf("service not found")
)
