package podwatcher

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
)

// Run start to run ctrlplane to watch
func Run(client kubernetes.Interface, stop chan struct{}) error {
	// run local ip ctrlplane
	if err := runLocalPodController(client, stop); err != nil {
		return fmt.Errorf("run local ip ctrlplane error: %v", err)
	}

	return nil
}
