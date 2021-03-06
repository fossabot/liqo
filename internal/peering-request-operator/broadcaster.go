package peering_request_operator

import (
	"context"
	discoveryv1 "github.com/liqoTech/liqo/api/discovery/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

func (r *PeeringRequestReconciler) BroadcasterExists(request *discoveryv1.PeeringRequest) (bool, error) {
	_, err := r.crdClient.Client().AppsV1().Deployments(r.Namespace).Get(context.TODO(), "broadcaster-"+request.Name, metav1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		// does not exist
		return false, nil
	}
	if err != nil {
		// other errors
		return false, err
	}
	// already exists
	return true, nil
}

func GetBroadcasterDeployment(request *discoveryv1.PeeringRequest, nameSA string, namespace string, image string, clusterId string, gatewayPrivateIP string) appsv1.Deployment {
	args := []string{
		"--peering-request",
		request.Name,
		"--cluster-id",
		clusterId,
		"--gateway-private-ip",
		gatewayPrivateIP,
		"--service-account",
		nameSA, //TODO: using this SA, we pass to the foreign cluster a kubeconfig with the same permissions of the broadcaster deployment; if we want to pass a different one we have to forge it
	}

	deploy := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "broadcaster-" + request.Name,
			Namespace: namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "v1",
					Kind:       "PeeringRequest",
					Name:       request.Name,
					UID:        request.UID,
				},
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "broadcaster-" + request.Name,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "broadcaster-" + request.Name,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            "broadcaster-" + request.Name,
							Image:           image,
							ImagePullPolicy: v1.PullAlways,
							Args:            args,
							Env: []v1.EnvVar{
								{
									Name:  "POD_NAMESPACE",
									Value: namespace,
								},
							},
						},
					},
					ServiceAccountName: nameSA,
				},
			},
		},
	}

	return deploy
}
