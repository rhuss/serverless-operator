package client

import (
	"context"

	mf "github.com/manifestival/manifestival"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewManifest(pathname string, client client.Client, opts ...mf.Option) (mf.Manifest, error) {
	return mf.NewManifest(pathname, append(opts, mf.UseClient(NewClient(client)))...)
}

func NewClient(client client.Client) mf.Client {
	return &controllerRuntimeClient{client: client}
}

type controllerRuntimeClient struct {
	client client.Client
}

// verify implementation
var _ mf.Client = (*controllerRuntimeClient)(nil)

func (c *controllerRuntimeClient) Create(obj *unstructured.Unstructured, options ...mf.ApplyOption) error {
	return c.client.Create(context.TODO(), obj)
}

func (c *controllerRuntimeClient) Update(obj *unstructured.Unstructured, options ...mf.ApplyOption) error {
	return c.client.Update(context.TODO(), obj)
}

func (c *controllerRuntimeClient) Delete(obj *unstructured.Unstructured, options ...mf.DeleteOption) error {
	err := c.client.Delete(context.TODO(), obj, deleteWith(options)...)
	if apierrors.IsNotFound(err) {
		opts := mf.DeleteWith(options)
		if opts.IgnoreNotFound {
			return nil
		}
	}
	return err
}

func (c *controllerRuntimeClient) Get(obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	key := client.ObjectKey{Namespace: obj.GetNamespace(), Name: obj.GetName()}
	result := &unstructured.Unstructured{}
	result.SetGroupVersionKind(obj.GroupVersionKind())
	err := c.client.Get(context.TODO(), key, result)
	return result, err
}

func deleteWith(opts []mf.DeleteOption) []client.DeleteOptionFunc {
	result := []client.DeleteOptionFunc{}
	for _, opt := range opts {
		switch v := opt.(type) {
		case mf.GracePeriodSeconds:
			result = append(result, client.GracePeriodSeconds(int64(v)))
		case mf.Preconditions:
			p := metav1.Preconditions(v)
			result = append(result, client.Preconditions(&p))
		case mf.PropagationPolicy:
			result = append(result, client.PropagationPolicy(metav1.DeletionPropagation(v)))
		}
	}
	return result
}
