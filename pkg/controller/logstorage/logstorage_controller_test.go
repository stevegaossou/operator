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
	BeforeEach(func() {
		c, mgr = setupManager()
		//ns := &corev1.Namespace{
		//	TypeMeta:   metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
		//	ObjectMeta: metav1.ObjectMeta{Name: "tigera-operator"},
		//	Spec:       corev1.NamespaceSpec{},
		//}
		//err := c.Create(context.Background(), ns)
		//if err != nil && !kerror.IsAlreadyExists(err) {
		//	Expect(err).NotTo(HaveOccurred())
		//}
	})
	//var c client.Client
	//BeforeEach(func() {
	//	// Create a Kubernetes client.
	//	cfg, err := config.GetConfig()
	//	Expect(err).NotTo(HaveOccurred())
	//	c, err = client.New(cfg, client.Options{})
	//	Expect(err).NotTo(HaveOccurred())
	//})

	It("should query a default LogStorage instance", func() {
		By("Creating a CRD")
		instance := &operatorv1.LogStorage{
			ObjectMeta: metav1.ObjectMeta{Name: "tigera-secure"},
		}
		err := c.Create(context.Background(), instance)
		if err != nil && !kerror.IsAlreadyExists(err) {
			Expect(err).NotTo(HaveOccurred())
		}
		_, err = GetLogStorage(context.Background(), c)
		Expect(err).NotTo(HaveOccurred())

		By("Running the operator")
		stopChan := RunOperator(mgr)
		defer close(stopChan)
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

func RunOperator(mgr manager.Manager) chan struct{} {
	stopChan := make(chan struct{})
	go func() {
		defer GinkgoRecover()
		err := mgr.Start(stopChan)
		Expect(err).NotTo(HaveOccurred())
	}()
	return stopChan
}

