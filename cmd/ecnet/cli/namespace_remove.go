package main

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"

	"github.com/flomesh-io/ErieCanal/pkg/ecnet/constants"
)

const namespaceRemoveDescription = `
This command will remove a namespace from the ecnet. All
services in this namespace will be removed from the ecnet.
`

type namespaceRemoveCmd struct {
	out       io.Writer
	namespace string
	ecnetName string
	clientSet kubernetes.Interface
}

func newNamespaceRemove(out io.Writer) *cobra.Command {
	namespaceRemove := &namespaceRemoveCmd{
		out: out,
	}

	cmd := &cobra.Command{
		Use:   "remove <NAMESPACE>",
		Short: "remove namespace from ecnet",
		Long:  namespaceRemoveDescription,
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			namespaceRemove.namespace = args[0]
			config, err := settings.RESTClientGetter().ToRESTConfig()
			if err != nil {
				return fmt.Errorf("Error fetching kubeconfig: %w", err)
			}

			clientset, err := kubernetes.NewForConfig(config)
			if err != nil {
				return fmt.Errorf("Could not access Kubernetes cluster, check kubeconfig: %w", err)
			}
			namespaceRemove.clientSet = clientset
			return namespaceRemove.run()
		},
	}

	//add ecnet name flag
	f := cmd.Flags()
	f.StringVar(&namespaceRemove.ecnetName, "ecnet-name", "ecnet", "Name of the ecnet")

	return cmd
}

func (r *namespaceRemoveCmd) run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	namespace, err := r.clientSet.CoreV1().Namespaces().Get(ctx, r.namespace, metav1.GetOptions{})

	if err != nil {
		return fmt.Errorf("Could not get namespace [%s]: %w", r.namespace, err)
	}

	val, exists := namespace.ObjectMeta.Labels[constants.ECNETKubeResourceMonitorAnnotation]

	if exists {
		if val == r.ecnetName {
			// Setting null for a key in a map removes only that specific key, which is the desired behavior.
			// Even if the key does not exist, there will be no side effects with setting the key to null, which
			// will result in the same behavior as if the key were present - the key being removed.
			patch := fmt.Sprintf(`
{
	"metadata": {
		"labels": {
			"%s": null,
			"%s": null
		},
		"annotations": {
			"%s": null,
			"%s": null
		}
	}
}`, constants.ECNETKubeResourceMonitorAnnotation, constants.IgnoreLabel, constants.SidecarInjectionAnnotation, constants.MetricsAnnotation)

			_, err = r.clientSet.CoreV1().Namespaces().Patch(ctx, r.namespace, types.StrategicMergePatchType, []byte(patch), metav1.PatchOptions{}, "")

			if err != nil {
				return fmt.Errorf("Could not remove namespace [%s] from ecnet [%s]: %w", r.namespace, r.ecnetName, err)
			}

			fmt.Fprintf(r.out, "Namespace [%s] successfully removed from ecnet [%s]\n", r.namespace, r.ecnetName)
		} else {
			return fmt.Errorf("Namespace belongs to ecnet [%s], not ecnet [%s]. Please specify the correct ecnet", val, r.ecnetName)
		}
	} else {
		fmt.Fprintf(r.out, "Namespace [%s] already does not belong to any ecnet\n", r.namespace)
		return nil
	}

	return nil
}
