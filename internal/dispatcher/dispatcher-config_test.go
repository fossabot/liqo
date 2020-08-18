package dispatcher

import (
	policyv1 "github.com/liqoTech/liqo/api/cluster-config/v1"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"
)

func TestDispatcherReconciler_GetConfig(t *testing.T) {
	dispatcher := DispatcherReconciler{}
	//test 1
	//the list of the resources to be replicated is 0, so we expect a 0 length list to be returned by the function
	t1 := policyv1.DispatcherConfig{ResourcesToReplicate: nil}
	//test 2
	//the list of the resources to be replicated contains 2 elements, so we expect  two elements in the list to be returned by the function
	t2 := policyv1.DispatcherConfig{ResourcesToReplicate: []policyv1.Resource{
		{
			Group:    "liqonet.liqo.io",
			Version:  "v1alpha1",
			Resource: "networkconfigs"},
		{
			Group:    "liqonet.liqo.io",
			Version:  "v1",
			Resource: "tunnelendpoints",
		},
	}}
	tests := []struct {
		config           policyv1.DispatcherConfig
		expectedElements int
	}{
		{t1, 0},
		{t2, 2},
	}

	for _, test := range tests {
		cfg := &policyv1.ClusterConfig{
			Spec: policyv1.ClusterConfigSpec{
				DispatcherConfig: test.config,
			},
			Status: policyv1.ClusterConfigStatus{},
		}
		res := dispatcher.GetConfig(cfg)
		assert.Equal(t, test.expectedElements, len(res), "length should be equal")
	}
}

func TestDispatcherReconciler_GetRemovedResources(t *testing.T) {
	dispatcher := DispatcherReconciler{
		RegisteredResources: []schema.GroupVersionResource{
			{
				Group:    "liqonet.liqo.io",
				Version:  "v1alpha1",
				Resource: "networkconfigs",
			},
			{
				Group:    "liqonet.liqo.io",
				Version:  "v1",
				Resource: "tunnelendpoints",
			},
		},
		Started: false,
	}
	//test 1
	//the configuration does not change, is the same
	//so we expect expect to get a 0 length list
	t1 := []schema.GroupVersionResource{
		{
			Group:    "liqonet.liqo.io",
			Version:  "v1alpha1",
			Resource: "networkconfigs"},
		{
			Group:    "liqonet.liqo.io",
			Version:  "v1",
			Resource: "tunnelendpoints",
		},
	}
	//test2
	//we remove a resource from the configuration and add a new one to it
	//so we expect to get a list with 1 element
	t2 := []schema.GroupVersionResource{
		{
			Group:    "liqonet.liqo.io",
			Version:  "v1alpha1",
			Resource: "networkconfigs"},
		{
			Group:    "liqonet.liqo.io",
			Version:  "v1alpha1",
			Resource: "tunnelendpoints",
		},
	}

	tests := []struct {
		config           []schema.GroupVersionResource
		expectedElements int
	}{
		{t1, 0},
		{t2, 1},
	}

	for _, test := range tests {
		res := dispatcher.GetRemovedResources(test.config)
		assert.Equal(t, test.expectedElements, len(res), "length should be equal")
	}
}

func TestDispatcherReconciler_UpdateConfig(t *testing.T) {
	dispatcher := DispatcherReconciler{}
	//test 1
	//the list of the resources to be replicated is 0, so we expect a 0 length list to be returned by the function
	//and 0 elements removed
	t1 := policyv1.DispatcherConfig{ResourcesToReplicate: nil}
	//test 2
	//the list of the resources to be replicated contains 2 elements, so we expect  two elements in the list to be returned by the function
	//and 0 elements removed
	t2 := policyv1.DispatcherConfig{ResourcesToReplicate: []policyv1.Resource{
		{
			Group:    "liqonet.liqo.io",
			Version:  "v1alpha1",
			Resource: "networkconfigs"},
		{
			Group:    "liqonet.liqo.io",
			Version:  "v1",
			Resource: "tunnelendpoints",
		},
	}}

	//test 3
	//we remove an existing element and add a new one. we expect to have 2 elements in the registeredResources
	//and 1 element removedResources
	t3 := policyv1.DispatcherConfig{ResourcesToReplicate: []policyv1.Resource{
		{
			Group:    "liqonet.liqo.io",
			Version:  "v1alpha1",
			Resource: "networkconfigs"},
		{
			Group:    "liqonet.liqo.io",
			Version:  "v1alpha1",
			Resource: "tunnelendpoints",
		},
	}}
	tests := []struct {
		config                     policyv1.DispatcherConfig
		expectedElementsResources  int
		expectedElementsRemovedRes int
	}{
		{t1, 0, 0},
		{t2, 2, 0},
		{t3, 2, 1},
	}

	for _, test := range tests {
		cfg := &policyv1.ClusterConfig{
			Spec: policyv1.ClusterConfigSpec{
				DispatcherConfig: test.config,
			},
			Status: policyv1.ClusterConfigStatus{},
		}
		dispatcher.UpdateConfig(cfg)
		assert.Equal(t, test.expectedElementsResources, len(dispatcher.RegisteredResources), "length should be equal")
		assert.Equal(t, test.expectedElementsRemovedRes, len(dispatcher.UnregisteredResources), "length should be equal")
	}
}

//we test that if the *rest.config of the custer is not correct the function return the error
func TestDispatcherReconciler_WatchConfiguration(t *testing.T) {
	dispatcher := DispatcherReconciler{}
	//test1
	//the group version is not correct and we expect an error
	config := k8sManagerLocal.GetConfig()
	err := dispatcher.WatchConfiguration(config, nil)
	assert.NotNil(t, err, "error should be not nil")

	//test2
	//the group version is not correct and we expect an error
	err = dispatcher.WatchConfiguration(config, &policyv1.GroupVersion)
	assert.Nil(t, err, "error should be not nil")
}
