package logstorage

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/operator-framework/operator-sdk/pkg/restmapper"
	"github.com/tigera/operator/pkg/apis"
	operatorv1 "github.com/tigera/operator/pkg/apis/operator/v1"
	kerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var _ = Describe("LogStorage controller tests", func() {
	var c client.Client
	var mgr manager.Manager
	//var reconciler reconcile.Reconciler
	BeforeEach(func() {
		c, mgr = setupManager()
	})

	It("should query a default LogStorage instance", func() {
		By("Creating a CRD")
		instance := &operatorv1.LogStorage{
			TypeMeta:   metav1.TypeMeta{Kind: "LogStorage", APIVersion: "operator.tigera.io/v1"},
			ObjectMeta: metav1.ObjectMeta{Name: "tigera-secure", SelfLink: "/apis/operator.tigera.io/v1/logstorages/tigera-secure"},
			Spec: operatorv1.LogStorageSpec{
				Nodes: &operatorv1.Nodes{
					Count: 1,
				},
			},
		}
		err := c.Create(context.Background(), instance)
		if err != nil && !kerror.IsAlreadyExists(err) {
			Expect(err).NotTo(HaveOccurred())
		}

		// add new logstorage controller to manager
		err = Add(mgr, operatorv1.ProviderNone, true)
		Expect(err).NotTo(HaveOccurred())
		_, err = GetLogStorage(context.Background(), c)
		Expect(err).NotTo(HaveOccurred())

		//// TODO: get rid of this before merging
		//By("Running the operator")
		//stopChan := RunOperator(mgr)
		//defer close(stopChan)
	})
})

func setupManager() (client.Client, manager.Manager) {
	// Create a Kubernetes client.
	cfg, err := config.GetConfig()
	Expect(err).NotTo(HaveOccurred())
	// Create a manager to use in the tests.
	mgr, err := manager.New(cfg, manager.Options{
		Namespace:      "",
		MapperProvider: restmapper.NewDynamicRESTMapper,
	})
	Expect(err).NotTo(HaveOccurred())
	// Setup Scheme for all resources
	err = apis.AddToScheme(mgr.GetScheme())
	Expect(err).NotTo(HaveOccurred())
	return mgr.GetClient(), mgr
}
//
//func RunOperator(mgr manager.Manager) chan struct{} {
//	stopChan := make(chan struct{})
//	go func() {
//		defer GinkgoRecover()
//		err := mgr.Start(stopChan)
//		Expect(err).NotTo(HaveOccurred())
//	}()
//	return stopChan
//}

