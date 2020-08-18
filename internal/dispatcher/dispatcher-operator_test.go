package dispatcher

import (
	"context"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"testing"
	"time"
)

func getObj() *unstructured.Unstructured {
	networkConfig := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "liqonet.liqo.io/v1alpha1",
			"kind":       "NetworkConfig",
			"metadata": map[string]interface{}{
				"name": "test-networkconfig",
			},
			"spec": map[string]interface{}{
				"clusterID":       "clusterID-test",
				"podCIDR":         "10.0.0.0/12",
				"tunnelPublicIP":  "192.16.5.1",
				"tunnelPrivateIP": "192.168.4.1",
			},
		},
	}
	return networkConfig
}

func TestDispatcherReconciler_CreateResource(t *testing.T) {
	networkConfig := getObj()
	d := DispatcherReconciler{}
	//test 1
	//the resource does not exist on the cluster
	//we expect to be created
	err := d.CreateResource(dynClient, gvr, networkConfig, clusterID)
	assert.Nil(t, err, "error should be nil")
	//test 2
	//the resource exists on the cluster and is the same
	//we expect not to be created and returns nil
	err = d.CreateResource(dynClient, gvr, networkConfig, clusterID)
	assert.Nil(t, err, "error should be nil")
	//test 3
	//the resource has different values than the existing one
	//we expect for the resource to be deleted and recreated
	networkConfig.SetLabels(map[string]string{"labelTestin": "test"})
	err = d.CreateResource(dynClient, gvr, networkConfig, clusterID)
	assert.Nil(t, err, "error should be nil")
	//test 4
	//the resource is not a valid one
	//we expect an error
	networkConfig.SetAPIVersion("invalidOne")
	networkConfig.SetLabels(map[string]string{"labelTesti": "test"})
	err = d.CreateResource(dynClient, gvr, networkConfig, clusterID)
	assert.NotNil(t, err, "error should not be nil")
	//test 5
	//the resource schema is not correct
	//we expect an error
	err = d.CreateResource(dynClient, schema.GroupVersionResource{}, networkConfig, clusterID)
	assert.NotNil(t, err, "error should not be nil")

}

func TestDispatcherReconciler_DeleteResource(t *testing.T) {
	d := DispatcherReconciler{}
	//test 1
	//delete an existing resource
	//we expect the error to be nil
	networkConfig := getObj()
	err := d.CreateResource(dynClient, gvr, networkConfig, clusterID)
	assert.Nil(t, err, "error should be nil")
	err = d.DeleteResource(dynClient, gvr, networkConfig, clusterID)
	assert.Nil(t, err, "error should be nil")
	//test 2
	//deleting a resource that does not exist
	//we expect an error
	err = d.DeleteResource(dynClient, gvr, networkConfig, clusterID)
	assert.NotNil(t, err, "error should be not nil")
}

func TestDispatcherReconciler_UpdateResource(t *testing.T) {
	d := DispatcherReconciler{}
	//first we create the resource
	networkConfig := getObj()
	err := d.CreateResource(dynClient, gvr, networkConfig, clusterID)
	assert.Nil(t, err, "error should be nil")

	//Test 1
	//we update the metadata section
	//we expect a nil error and the metadata section of the resource on the server to be equal
	networkConfig.SetLabels(map[string]string{"labelTesting": "test"})
	err = d.UpdateResource(dynClient, gvr, networkConfig, clusterID)
	assert.Nil(t, err, "error should be nil")
	obj, err := dynClient.Resource(gvr).Get(context.TODO(), networkConfig.GetName(), metav1.GetOptions{})
	assert.Nil(t, err, "error should be nil")
	assert.Equal(t, networkConfig.GetLabels(), obj.GetLabels(), "labels should be equal")

	//Test 2
	//we update the spec section
	//we expect a nil error and the spec section of the resource on the server to be equal as we set it
	newSpec, err := getSpec(networkConfig, clusterID)
	assert.Nil(t, err, "error should be nil")
	newSpec["podCIDR"] = "1.1.1.1"
	//setting the new values of spec fields
	err = unstructured.SetNestedMap(obj.Object, newSpec, "spec")
	assert.Nil(t, err, "error should be nil")
	err = d.UpdateResource(dynClient, gvr, obj, clusterID)
	assert.Nil(t, err, "error should be nil")
	obj, err = dynClient.Resource(gvr).Get(context.TODO(), networkConfig.GetName(), metav1.GetOptions{})
	assert.Nil(t, err, "error should be nil")
	spec, err := getSpec(obj, clusterID)
	assert.Nil(t, err, "error should be nil")
	assert.Equal(t, newSpec, spec, "specs should be equal")

	//Test 3
	//we update the status section
	//we expect a nil error and the status section of the resource on the server to be equal as we set it
	newStatus := map[string]interface{}{
		"natEnabled": true,
	}
	err = unstructured.SetNestedMap(obj.Object, newStatus, "status")
	assert.Nil(t, err, "error should be nil")
	err = d.UpdateResource(dynClient, gvr, obj, clusterID)
	assert.Nil(t, err, "error should be nil")
	obj, err = dynClient.Resource(gvr).Get(context.TODO(), networkConfig.GetName(), metav1.GetOptions{})
	assert.Nil(t, err, "error should be nil")
	status, err := getStatus(obj, clusterID)
	assert.Nil(t, err, "error should be nil")
	assert.Equal(t, newStatus, status, "status should be equal")
}

func TestDispatcherReconciler_StartWatchers(t *testing.T) {
	d := DispatcherReconciler{
		LocalDynClient:  dynClient,
		RunningWatchers: make(map[string]chan bool),
		Started:         false,
	}
	//for each test we have a number of registers resources and
	//after calling the StartWatchers function we expect two have a certain number of active watchers
	//as is the number of the registered resources
	test1 := []schema.GroupVersionResource{{
		Group:    "liqonet.liqo.io",
		Version:  "v1alpha1",
		Resource: "networkconfigs",
	}, {
		Group:    "liqonet.liqo.io",
		Version:  "v1",
		Resource: "tunnelendpoints",
	}}
	test2 := []schema.GroupVersionResource{}
	test3 := []schema.GroupVersionResource{{
		Group:    "liqonet.liqo.io",
		Version:  "v1alpha1",
		Resource: "networkconfigs",
	}}
	tests := []struct {
		test             []schema.GroupVersionResource
		expectedWatchers int
	}{
		{test1, 2},
		{test2, 0},
		{test3, 1},
	}

	for _, test := range tests {
		d.RegisteredResources = test.test
		d.StartWatchers()
		assert.Equal(t, test.expectedWatchers, len(d.RunningWatchers), "it should be the same")
		//stop the watchers
		for k, ch := range d.RunningWatchers {
			close(ch)
			delete(d.RunningWatchers, k)
			time.Sleep(1 * time.Second)
		}
	}
	//test on a closed channel
	//we close a channel of a running watcher an expect that the function restarts the watcher
	//we add a new channel on runningWatchers
	d.RunningWatchers[test3[0].String()] = make(chan bool)
	close(d.RunningWatchers[test3[0].String()])
	time.Sleep(1 * time.Second)
	d.StartWatchers()
	select {
	case _, ok := <-d.RunningWatchers[test3[0].String()]:
		assert.True(t, ok, "should be true")
	default:

	}
	assert.NotPanics(t, func() { close(d.RunningWatchers[test3[0].String()]) }, "should not panic")
}

func TestDispatcherReconciler_StopWatchers(t *testing.T) {
	d := DispatcherReconciler{
		LocalDynClient:  dynClient,
		RunningWatchers: make(map[string]chan bool),
		Started:         false,
	}
	//we add two kind of resources to be watched
	//then unregister them and check that the watchers have ben closed as well
	test1 := []schema.GroupVersionResource{{
		Group:    "liqonet.liqo.io",
		Version:  "v1alpha1",
		Resource: "networkconfigs",
	}, {
		Group:    "liqonet.liqo.io",
		Version:  "v1",
		Resource: "tunnelendpoints",
	}}
	d.RegisteredResources = test1
	d.StartWatchers()
	assert.Equal(t, 2, len(d.RunningWatchers), "it should be 2")
	for _, r := range test1 {
		d.UnregisteredResources = append(d.UnregisteredResources, r.String())
	}
	d.StopWatchers()
	assert.Equal(t, 0, len(d.RunningWatchers), "it should be 0")
	d.UnregisteredResources = []string{}
	//test 2
	//we close previously a channel of a watcher and then we add the resource to the unregistered list
	//we expect than it does not panic and only one watcher is still active
	d.RegisteredResources = test1
	d.StartWatchers()
	assert.Equal(t, 2, len(d.RunningWatchers), "it should be 2")
	d.UnregisteredResources = append(d.UnregisteredResources, d.RegisteredResources[0].String())
	assert.NotPanics(t, func() { close(d.RunningWatchers[d.RegisteredResources[0].String()]) }, "should not panic")
	d.StopWatchers()
	assert.Equal(t, 1, len(d.RunningWatchers), "it should be 0")
}

func TestDispatcherReconciler_AddedHandler(t *testing.T) {
	d := DispatcherReconciler{
		RemoteDynClients: map[string]dynamic.Interface{clusterID: dynClient},
	}
	//test 1
	//adding a resource kind that exists on the cluster
	//we expect the resource to be created
	test1 := getObj()
	d.AddedHandler(test1, gvr)
	time.Sleep(1 * time.Second)
	obj, err := dynClient.Resource(gvr).Get(context.TODO(), test1.GetName(), metav1.GetOptions{})
	assert.Nil(t, err, "error should be empty")
	assert.True(t, areEqual(test1, obj), "the two objects should be equal")

	//remove the resource
	err = dynClient.Resource(gvr).Delete(context.TODO(), test1.GetName(), metav1.DeleteOptions{})
	assert.Nil(t, err, "should be nil")

	//test 2
	//adding a resource kind that the api server does not know
	//we expect an error to be returned
	d.AddedHandler(test1, schema.GroupVersionResource{})
	obj, err = dynClient.Resource(gvr).Get(context.TODO(), test1.GetName(), metav1.GetOptions{})
	assert.NotNil(t, err, "error should be not nil")
	assert.Nil(t, obj, "the object retrieved should be nil")
}
func TestDispatcherReconciler_ModifiedHandler(t *testing.T) {
	d := DispatcherReconciler{
		RemoteDynClients: map[string]dynamic.Interface{clusterID: dynClient},
	}

	//test 1
	//the modified resource does not exist on the cluster
	//we expect the resource to be created and error to be nil
	test1 := getObj()
	d.ModifiedHandler(test1, gvr)
	time.Sleep(1 * time.Second)
	obj, err := dynClient.Resource(gvr).Get(context.TODO(), test1.GetName(), metav1.GetOptions{})
	assert.Nil(t, err, "error should be empty")
	assert.True(t, areEqual(test1, obj), "the two objects should be equal")

	//test 2
	//the modified resource already exists on the cluster
	//we expect the resource to be modified and the error to be nil
	test1.SetLabels(map[string]string{
		"labelTestin": "labelling",
	})
	d.ModifiedHandler(test1, gvr)
	time.Sleep(1 * time.Second)
	obj, err = dynClient.Resource(gvr).Get(context.TODO(), test1.GetName(), metav1.GetOptions{})
	assert.Nil(t, err, "error should be empty")
	assert.True(t, areEqual(test1, obj), "the two objects should be equal")

	//clean up the resource
	err = dynClient.Resource(gvr).Delete(context.TODO(), test1.GetName(), metav1.DeleteOptions{})
	assert.Nil(t, err, "should be nil")
}

func TestDispatcherReconciler_DeletedHandler(t *testing.T) {
	d := DispatcherReconciler{
		RemoteDynClients: map[string]dynamic.Interface{clusterID: dynClient},
	}
	//test 1
	//we create a resource then we pass it to the handler
	//we expect the resource to be deleted
	test1 := getObj()
	obj, err := dynClient.Resource(gvr).Create(context.TODO(), test1, metav1.CreateOptions{})
	assert.Nil(t, err, "error should be nil")
	assert.True(t, areEqual(test1, obj), "the two objects should be equal")
	d.DeletedHandler(obj, gvr)
	obj, err = dynClient.Resource(gvr).Get(context.TODO(), test1.GetName(), metav1.GetOptions{})
	assert.NotNil(t, err, "error should not be empty")
	assert.Nil(t, obj, "the object retrieved should be nil")
}

func TestIsOpen(t *testing.T) {
	//test 1
	//create a bool channel
	//expect to be opened
	ch := make(chan bool, 1)
	result := isOpen(ch)
	assert.True(t, result, "channel should be open")

	//test 2
	//write to the channel
	//expect to be opened
	ch <- true
	result = isOpen(ch)
	assert.True(t, result, "channel should be open")

	//test 3
	//close the channel and check if is closed
	//expect to be closed
	assert.NotPanics(t, func() { close(ch) }, "this should not panic, because the channel is opened")
	result = isOpen(ch)
	assert.False(t, result, "channel should be closed")

}

func TestGetSpec(t *testing.T) {
	spec := map[string]interface{}{
		"clusterID": "clusterID-test",
	}
	//test 1
	//we have an object with a spec field
	//we expect to get the spec and a nil error
	test1 := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec": spec,
		},
	}
	objSpec, err := getSpec(test1, clusterID)
	assert.Nil(t, err, "error should be nil")
	assert.Equal(t, spec, objSpec, "the two specs should be equal")

	//test 2
	//we have an object without a spec field
	//we expect the error to be and a nil spec to be returned because the specied field is not found
	test2 := &unstructured.Unstructured{
		Object: map[string]interface{}{},
	}
	objSpec, err = getSpec(test2, clusterID)
	assert.Nil(t, err, "error should be nil")
	assert.Nil(t, objSpec, "the spec should be nil")
}

func TestGetStatus(t *testing.T) {
	status := map[string]interface{}{
		"natEnabled": true,
	}
	//test 1
	//we have an object with a status field
	//we expect to get the status and a nil error
	test1 := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"status": status,
		},
	}
	objStatus, err := getStatus(test1, clusterID)
	assert.Nil(t, err, "error should be nil")
	assert.Equal(t, status, objStatus, "the two specs should be equal")

	//test 2
	//we have an object without a spec field
	//we expect the error to be and a nil spec to be returned because the specied field is not found
	test2 := &unstructured.Unstructured{
		Object: map[string]interface{}{},
	}
	objStatus, err = getStatus(test2, clusterID)
	assert.Nil(t, err, "error should be nil")
	assert.Nil(t, objStatus, "the spec should be nil")
}
