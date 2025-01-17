package clusterconnection_test

import (
	"context"

	"github.com/tigera/operator/pkg/controller/clusterconnection"
	"github.com/tigera/operator/pkg/render"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/tigera/operator/pkg/apis"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	operatorv1 "github.com/tigera/operator/pkg/apis/operator/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("ManagementClusterConnection controller tests", func() {
	var c client.Client
	var ctx context.Context
	var cfg *operatorv1.ManagementClusterConnection
	var r clusterconnection.ReconcileConnection
	var scheme *runtime.Scheme
	var dpl *appsv1.Deployment

	BeforeSuite(func() {
		// Create a Kubernetes client.
		scheme = runtime.NewScheme()
		err := apis.AddToScheme(scheme)
		Expect(err).NotTo(HaveOccurred())
		scheme.AddKnownTypes(schema.GroupVersion{Group: "apps", Version: "v1"}, &appsv1.Deployment{})
		scheme.AddKnownTypes(schema.GroupVersion{Group: "", Version: "v1"}, &rbacv1.ClusterRole{})
		scheme.AddKnownTypes(schema.GroupVersion{Group: "", Version: "v1"}, &rbacv1.ClusterRoleBinding{})
		err = operatorv1.SchemeBuilder.AddToScheme(scheme)
		Expect(err).NotTo(HaveOccurred())
		c = fake.NewFakeClientWithScheme(scheme)
		ctx = context.Background()
		r = clusterconnection.ReconcileConnection{
			Client:   c,
			Scheme:   scheme,
			Provider: operatorv1.ProviderNone,
		}
		dpl = &appsv1.Deployment{
			TypeMeta: metav1.TypeMeta{Kind: "Deployment", APIVersion: "apps/v1"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      render.GuardianDeploymentName,
				Namespace: render.GuardianNamespace,
			},
		}
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      render.GuardianSecretName,
				Namespace: render.OperatorNamespace(),
			},
			Data: map[string][]byte{
				"cert": []byte("foo"),
				"key":  []byte("bar"),
			},
		}
		c.Create(ctx, secret)
	})

	It("should create a default ManagementClusterConnection", func() {

		By("applying the required prerequisites")
		// Create a ManagementClusterConnection in the k8s client.
		cfg = &operatorv1.ManagementClusterConnection{
			ObjectMeta: metav1.ObjectMeta{Name: "tigera-secure"},
			Spec: operatorv1.ManagementClusterConnectionSpec{
				ManagementClusterAddr: "127.0.0.1:12345",
			},
		}
		err := c.Create(ctx, cfg)
		Expect(err).NotTo(HaveOccurred())
		err = c.Create(
			ctx,
			&operatorv1.Installation{
				Spec: operatorv1.InstallationSpec{
					Registry: "my-reg",
					// The test is provider agnostic.
					KubernetesProvider: operatorv1.ProviderNone,
				},
				ObjectMeta: metav1.ObjectMeta{Name: "default"},
			})
		Expect(err).NotTo(HaveOccurred())
		By("reconciling with the required prerequisites")
		err = c.Get(ctx, client.ObjectKey{Name: render.GuardianDeploymentName, Namespace: render.GuardianNamespace}, dpl)
		Expect(err).To(HaveOccurred())
		_, err = r.Reconcile(reconcile.Request{})
		err = c.Get(ctx, client.ObjectKey{Name: render.GuardianDeploymentName, Namespace: render.GuardianNamespace}, dpl)
		// Verifying that there is a deployment is enough for the purpose of this test. More detailed testing will be done
		// in the render package.
		Expect(err).NotTo(HaveOccurred())
		Expect(dpl.Labels["k8s-app"]).To(Equal(render.GuardianName))

	})
})
