package dispatcher

import (
	policyv1 "github.com/liqoTech/liqo/api/cluster-config/v1"
	"github.com/liqoTech/liqo/pkg/clusterConfig"
	"github.com/liqoTech/liqo/pkg/crdClient"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"reflect"
)

func (d *DispatcherReconciler) WatchConfiguration(config *rest.Config, gv *schema.GroupVersion) error {
	config.ContentConfig.GroupVersion = gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	config.UserAgent = rest.DefaultKubernetesUserAgent()
	CRDclient, err := crdClient.NewFromConfig(config)
	if err != nil {
		klog.Errorf("an error occurred while starting the watcher for the clusterConfig CRD: %s", err)
		return err
	}
	go clusterConfig.WatchConfiguration(d.UpdateConfig, CRDclient, "")
	return nil
}

func (d *DispatcherReconciler) UpdateConfig(cfg *policyv1.ClusterConfig) {
	resources := d.GetConfig(cfg)
	if !reflect.DeepEqual(d.RegisteredResources, resources) {
		klog.Info("updating the list of registered resources to be replicated")
		d.UnregisteredResources = d.GetRemovedResources(resources)
		d.RegisteredResources = resources
	}
}

func (d *DispatcherReconciler) GetConfig(cfg *policyv1.ClusterConfig) []schema.GroupVersionResource {
	resourceList := cfg.Spec.DispatcherConfig
	klog.Info(resourceList)
	config := []schema.GroupVersionResource{}
	for _, res := range resourceList.ResourcesToReplicate {
		config = append(config, schema.GroupVersionResource{
			Group:    res.Group,
			Version:  res.Version,
			Resource: res.Resource,
		})
	}
	return config
}

func (d *DispatcherReconciler) GetRemovedResources(resources []schema.GroupVersionResource) []string {
	oldRes := []string{}
	diffRes := []string{}
	newRes := []string{}
	//save the resources as strings in 'newRes'
	for _, r := range resources {
		newRes = append(newRes, r.String())
	}
	//get the old resources
	for _, r := range d.RegisteredResources {
		oldRes = append(oldRes, r.String())
	}
	//save in diffRes all the resources that appears in oldRes but not in newRes
	flag := false
	for _, old := range oldRes {
		for _, new := range newRes {
			if old == new {
				flag = true
				break
			}
		}
		if flag {
			flag = false
		} else {
			diffRes = append(diffRes, old)
		}
	}
	return diffRes
}
