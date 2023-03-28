/*
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
// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/flomesh-io/ErieCanal/pkg/ecnet/apis/config/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeEcnetConfigs implements EcnetConfigInterface
type FakeEcnetConfigs struct {
	Fake *FakeConfigV1alpha1
	ns   string
}

var ecnetconfigsResource = schema.GroupVersionResource{Group: "config.flomesh.io", Version: "v1alpha1", Resource: "ecnetconfigs"}

var ecnetconfigsKind = schema.GroupVersionKind{Group: "config.flomesh.io", Version: "v1alpha1", Kind: "EcnetConfig"}

// Get takes name of the ecnetConfig, and returns the corresponding ecnetConfig object, and an error if there is any.
func (c *FakeEcnetConfigs) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.EcnetConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(ecnetconfigsResource, c.ns, name), &v1alpha1.EcnetConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.EcnetConfig), err
}

// List takes label and field selectors, and returns the list of EcnetConfigs that match those selectors.
func (c *FakeEcnetConfigs) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.EcnetConfigList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(ecnetconfigsResource, ecnetconfigsKind, c.ns, opts), &v1alpha1.EcnetConfigList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.EcnetConfigList{ListMeta: obj.(*v1alpha1.EcnetConfigList).ListMeta}
	for _, item := range obj.(*v1alpha1.EcnetConfigList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested ecnetConfigs.
func (c *FakeEcnetConfigs) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(ecnetconfigsResource, c.ns, opts))

}

// Create takes the representation of a ecnetConfig and creates it.  Returns the server's representation of the ecnetConfig, and an error, if there is any.
func (c *FakeEcnetConfigs) Create(ctx context.Context, ecnetConfig *v1alpha1.EcnetConfig, opts v1.CreateOptions) (result *v1alpha1.EcnetConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(ecnetconfigsResource, c.ns, ecnetConfig), &v1alpha1.EcnetConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.EcnetConfig), err
}

// Update takes the representation of a ecnetConfig and updates it. Returns the server's representation of the ecnetConfig, and an error, if there is any.
func (c *FakeEcnetConfigs) Update(ctx context.Context, ecnetConfig *v1alpha1.EcnetConfig, opts v1.UpdateOptions) (result *v1alpha1.EcnetConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(ecnetconfigsResource, c.ns, ecnetConfig), &v1alpha1.EcnetConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.EcnetConfig), err
}

// Delete takes name of the ecnetConfig and deletes it. Returns an error if one occurs.
func (c *FakeEcnetConfigs) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(ecnetconfigsResource, c.ns, name, opts), &v1alpha1.EcnetConfig{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeEcnetConfigs) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(ecnetconfigsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.EcnetConfigList{})
	return err
}

// Patch applies the patch and returns the patched ecnetConfig.
func (c *FakeEcnetConfigs) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.EcnetConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(ecnetconfigsResource, c.ns, name, pt, data, subresources...), &v1alpha1.EcnetConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.EcnetConfig), err
}