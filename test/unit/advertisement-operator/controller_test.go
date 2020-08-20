package advertisement_operator

import (
	v1 "github.com/liqoTech/liqo/api/advertisement-operator/v1"
	policyv1 "github.com/liqoTech/liqo/api/cluster-config/v1"
	advcontroller "github.com/liqoTech/liqo/internal/advertisement-operator"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func createReconciler(acceptedAdv, maxAcceptableAdv int32, acceptPolicy policyv1.AcceptPolicy) advcontroller.AdvertisementReconciler {
	return advcontroller.AdvertisementReconciler{
		Client:           nil,
		Scheme:           nil,
		EventsRecorder:   nil,
		KubeletNamespace: "",
		KindEnvironment:  false,
		VKImage:          "",
		InitVKImage:      "",
		HomeClusterId:    "",
		AcceptedAdvNum:   acceptedAdv,
		ClusterConfig: policyv1.AdvertisementConfig{
			AdvOperatorConfig: policyv1.AdvOperatorConfig{
				MaxAcceptableAdvertisement: maxAcceptableAdv,
				AcceptPolicy:               acceptPolicy,
			},
		},
	}
}

func createAdvertisement() v1.Advertisement {
	return v1.Advertisement{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       v1.AdvertisementSpec{},
		Status:     v1.AdvertisementStatus{},
	}
}

func TestAutoAcceptAllAdvertisement(t *testing.T) {
	r := createReconciler(0, 10, policyv1.AutoAcceptAll)

	// given a configuration with AutoAcceptAll policy, create 15 Advertisements and check that they are all accepted, even if the Maximum is set to 10
	for i := 0; i < 15; i++ {
		adv := createAdvertisement()
		r.CheckAdvertisement(&adv)
		assert.Equal(t, advcontroller.AdvertisementAccepted, adv.Status.AdvertisementStatus)
	}
	// check that the Adv counter has been incremented
	assert.Equal(t, int32(15), r.AcceptedAdvNum)
}

func TestAutoAcceptAdvertisementWithinMaximum(t *testing.T) {
	r := createReconciler(0, 10, policyv1.AutoAcceptWithinMaximum)

	// given a configuration with max 10 Advertisements, create 10 Advertisements
	for i := 0; i < 10; i++ {
		adv := createAdvertisement()
		r.CheckAdvertisement(&adv)
	}

	// create 10 more Advertisements and check that they are all refused, since the maximum has been reached
	for i := 0; i < 10; i++ {
		adv := createAdvertisement()
		r.CheckAdvertisement(&adv)
		assert.Equal(t, advcontroller.AdvertisementRefused, adv.Status.AdvertisementStatus)
	}
	// check that the Adv counter has not been modified
	assert.Equal(t, int32(10), r.AcceptedAdvNum)
}

func TestAutoRefuseAdvertisement(t *testing.T) {
	r := createReconciler(0, 10, policyv1.AutoRefuseAll)

	// given a configuration with max 10 Advertisements but RefuseAll policy, create 10 Advertisements and check they are refused
	for i := 0; i < 10; i++ {
		adv := createAdvertisement()
		r.CheckAdvertisement(&adv)
		assert.Equal(t, advcontroller.AdvertisementRefused, adv.Status.AdvertisementStatus)
	}
	// check that the Adv counter has not been incremented
	assert.Equal(t, int32(0), r.AcceptedAdvNum)
}

func TestManageConfigUpdate(t *testing.T) {
	r := createReconciler(0, 10, policyv1.AutoAcceptWithinMaximum)
	advList := v1.AdvertisementList{
		Items: []v1.Advertisement{},
	}

	advCount := 15

	// given a configuration with max 10 Advertisements, create 15 Advertisement: 10 should be accepted and 5 refused
	for i := 0; i < advCount; i++ {
		adv := createAdvertisement()
		r.CheckAdvertisement(&adv)
		advList.Items = append(advList.Items, adv)
	}

	// the advList contains 10 accepted and 5 refused Adv
	// create a new configuration with MaxAcceptableAdv = 15
	// with the new configuration, check the 5 refused Adv are accepted
	config := policyv1.ClusterConfig{
		Spec: policyv1.ClusterConfigSpec{
			AdvertisementConfig: policyv1.AdvertisementConfig{
				AdvOperatorConfig: policyv1.AdvOperatorConfig{
					MaxAcceptableAdvertisement: int32(advCount),
					AcceptPolicy:               policyv1.AutoAcceptWithinMaximum,
				},
			},
		},
	}

	// TRUE TEST
	// test the true branch of ManageMaximumUpdate
	err, flag := r.ManageMaximumUpdate(config.Spec.AdvertisementConfig, &advList)
	assert.Nil(t, err)
	assert.True(t, flag)
	assert.Equal(t, config.Spec.AdvertisementConfig, r.ClusterConfig)
	assert.Equal(t, int32(advCount), r.AcceptedAdvNum)
	for _, adv := range advList.Items {
		assert.Equal(t, advcontroller.AdvertisementAccepted, adv.Status.AdvertisementStatus)
	}

	// FALSE TEST
	// apply again the same configuration
	// we enter in the false branch of ManageMaximumUpdate but nothing should change
	err, flag = r.ManageMaximumUpdate(config.Spec.AdvertisementConfig, &advList)
	assert.Nil(t, err)
	assert.False(t, flag)
	assert.Equal(t, config.Spec.AdvertisementConfig, r.ClusterConfig)
	assert.Equal(t, int32(advCount), r.AcceptedAdvNum)

	//TODO: FALSE TEST with config.MaxAcceptableAdvertisement < r.AcceptedAdvNum
	//      cannot test it yet (it needs a client)
}
