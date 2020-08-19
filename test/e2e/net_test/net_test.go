package net_test

import (
	"context"
	"github.com/liqoTech/liqo/test/e2e/util"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog"
	"testing"
	"time"
)

var (
	image           = "nginx"
	namespaceLabels = map[string]string{"liqo.io/enabled": "true"}
	waitTime        = 40 * time.Second
)

func TestPodConnectivity1to2(t *testing.T) {
	context := util.GetTester()
	NetworkParametersCheck(context, t)
	ConnectivityCheckPodToPod(context, t)
}

func NetworkParametersCheck(con *util.Tester, t *testing.T) {
	//create dynamic client for cluster 1
	dynClient1 := dynamic.NewForConfigOrDie(con.Config1)
	//get the Tunnel endpoint on cluster1
	tun1, err := dynClient1.Resource(schema.GroupVersionResource{
		Group:    "liqonet.liqo.io",
		Version:  "v1",
		Resource: "tunnelendpoints",
	}).Get(context.TODO(), con.ClusterID2+"-tunendpoint", metav1.GetOptions{})
	if err != nil {
		klog.Errorf("an error occurred while getting tunnelendpointCRD on cluster %s: %s", con.ClusterID1, err)
		t.Fail()
	}
	status, found, err := unstructured.NestedMap(tun1.Object, "status")
	if err != nil {
		klog.Errorf("an error occurred while getting status of tunnelendpointCRD on cluster %s: %s", con.ClusterID1, err)
		t.Fail()
	}
	if !found {
		klog.Infof("status is not yet setted for tunnelendpointCRD on cluster %s", con.ClusterID1)
	} else {
		klog.Infof("status of tunnelendpointCRD on cluster %s: %s", con.ClusterID1, status)
	}

	//create dynamic client for cluster 2
	dynClient2 := dynamic.NewForConfigOrDie(con.Config2)
	//get the Tunnel endpoint on cluster1
	tun2, err := dynClient2.Resource(schema.GroupVersionResource{
		Group:    "liqonet.liqo.io",
		Version:  "v1",
		Resource: "tunnelendpoints",
	}).Get(context.TODO(), con.ClusterID1+"-tunendpoint", metav1.GetOptions{})
	if err != nil {
		klog.Errorf("an error occurred while getting tunnelendpointCRD on cluster %s: %s", con.ClusterID1, err)
		t.Fail()
	}
	status, found, err = unstructured.NestedMap(tun2.Object, "status")
	if err != nil {
		klog.Errorf("an error occurred while getting status of tunnelendpointCRD on cluster %s: %s", con.ClusterID2, err)
		t.Fail()
	}
	if !found {
		klog.Infof("status is not yet setted for tunnelendpointCRD on cluster %s", con.ClusterID2)
	} else {
		klog.Infof("status of tunnelendpointCRD on cluster %s: %s", con.ClusterID2, status)
	}

}

func ConnectivityCheckPodToPod(con *util.Tester, t *testing.T) {
	ns := v1.Namespace{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "test-connectivity",
			Labels: namespaceLabels,
		},
		Spec:   v1.NamespaceSpec{},
		Status: v1.NamespaceStatus{},
	}
	reflectedNamespace := ns.Name + "-" + con.ClusterID1

	localNodes, err := con.Client1.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
		LabelSelector: "type!=virtual-node",
	})
	if err != nil {
		klog.Error(err)
		t.Fail()
	}
	remoteNodes, err := con.Client2.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
		LabelSelector: "type!=virtual-node",
	})
	if err != nil {
		klog.Error(err)
		t.Fail()
	}
	_, err = con.Client1.CoreV1().Namespaces().Create(context.TODO(), &ns, metav1.CreateOptions{})
	if err != nil {
		klog.Error(err)
		t.Fail()
	}
	podRemote := DeployRemotePod(image, ns.Name)
	_, err = con.Client1.CoreV1().Pods(ns.Name).Create(context.TODO(), podRemote, metav1.CreateOptions{})
	if err != nil {
		klog.Error(err)
		t.Fail()
	}

	podLocal := DeployLocalPod(image, ns.Name)
	_, err = con.Client1.CoreV1().Pods(ns.Name).Create(context.TODO(), podLocal, metav1.CreateOptions{})
	if err != nil {
		klog.Error(err)
		t.Fail()
	}

	if !util.WaitForPodToBeReady(con.Client1, waitTime, podLocal.Namespace, podLocal.Name) {
		klog.Infof("pod is not ready")
		t.Fail()
	}
	if !util.WaitForPodToBeReady(con.Client1, waitTime, podRemote.Namespace, podRemote.Name) {
		klog.Infof("pod is not ready")
		t.Fail()
	}
	if !util.WaitForPodToBeReady(con.Client2, waitTime, reflectedNamespace, podRemote.Name) {
		klog.Infof("pod is not ready")
		t.Fail()
	}
	podRemoteUpdateCluster2, err := con.Client2.CoreV1().Pods(reflectedNamespace).Get(context.TODO(), podRemote.Name, metav1.GetOptions{})
	if err != nil {
		klog.Error(err)
		t.Fail()
	}
	podRemoteUpdateCluster1, err := con.Client1.CoreV1().Pods(podRemote.Namespace).Get(context.TODO(), podRemote.Name, metav1.GetOptions{})
	if err != nil {
		klog.Error(err)
		t.Fail()
	}
	podLocalUpdate, err := con.Client1.CoreV1().Pods(podLocal.Namespace).Get(context.TODO(), podLocal.Name, metav1.GetOptions{})
	if err != nil {
		klog.Error(err)
		t.Fail()
	}
	assert.True(t, isContained(remoteNodes, podRemoteUpdateCluster2.Spec.NodeName), "remotepod should be running on one of the local nodes")
	assert.True(t, isContained(localNodes, podLocalUpdate.Spec.NodeName), "localpod should be running on one of the remote pods")
	cmd := "curl -s -o /dev/null -w '%{http_code}' " + podRemoteUpdateCluster1.Status.PodIP
	stdout, stderr, err := util.ExecCmd(con.Config1, con.Client1, podLocalUpdate.Name, podLocalUpdate.Namespace, cmd)
	assert.Equal(t, "200", stdout, "status code should be 200")
	if err != nil {
		klog.Error(err)
		klog.Infof(stdout)
		klog.Infof(stderr)
		t.Fail()
	}

	err = con.Client1.CoreV1().Namespaces().Delete(context.TODO(), ns.Name, metav1.DeleteOptions{})
	if err != nil {
		klog.Error(err)
		t.Fail()
	}
}

func DeployRemotePod(image, namespace string) *v1.Pod {
	pod1 := v1.Pod{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tester-remote",
			Namespace: namespace,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:            "tester",
					Image:           image,
					Resources:       v1.ResourceRequirements{},
					ImagePullPolicy: "IfNotPresent",
					Ports: []v1.ContainerPort{{
						ContainerPort: 80,
					}},
				},
			},
			Affinity: &v1.Affinity{
				NodeAffinity: &v1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{{
						MatchExpressions: []v1.NodeSelectorRequirement{{
							Key:      "type",
							Operator: "In",
							Values:   []string{"virtual-node"},
						}},
						MatchFields: nil,
					}}},
				},
			},
		},
		Status: v1.PodStatus{},
	}
	return &pod1
}

func DeployLocalPod(image, namespace string) *v1.Pod {
	pod2 := v1.Pod{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tester-local",
			Namespace: namespace,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:            "tester",
					Image:           image,
					ImagePullPolicy: "IfNotPresent",
					Ports: []v1.ContainerPort{{
						ContainerPort: 80,
					},
					},
				}},

			Affinity: &v1.Affinity{
				NodeAffinity: &v1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{{
						MatchExpressions: []v1.NodeSelectorRequirement{{
							Key:      "type",
							Operator: "NotIn",
							Values:   []string{"virtual-node"},
						}},
						MatchFields: nil,
					}}},
				},
			},
		},
	}
	return &pod2
}

func isContained(nodes *v1.NodeList, nodeName string) bool {
	for _, node := range nodes.Items {
		if nodeName == node.Name {
			return true
		}
	}
	return false
}
