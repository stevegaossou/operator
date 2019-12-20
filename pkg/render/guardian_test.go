package render_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tigera/operator/pkg/render"
	"k8s.io/apimachinery/pkg/runtime"

	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Rendering tests", func() {
	var g render.Component
	var resources []runtime.Object

	BeforeEach(func() {
		addr := "127.0.0.1:1234"

		g, _ = render.Guardian(
			addr,
			[]*corev1.Secret{},
			"cluster",
			false,
			"my-reg",
		)
		resources = g.Objects()
	})

	It("should render all resources for a managed cluster", func() {
		Expect(len(resources)).To(Equal(6))

		expectedResources := []struct {
			name    string
			ns      string
			group   string
			version string
			kind    string
		}{
			{name: render.GuardianNamespace, ns: "", group: "", version: "v1", kind: "Namespace"},
			{name: render.GuardianServiceAccountName, ns: render.GuardianNamespace, group: "", version: "v1", kind: "ServiceAccount"},
			{name: render.GuardianClusterRoleName, ns: "", group: "rbac.authorization.k8s.io", version: "v1", kind: "ClusterRole"},
			{name: render.GuardianClusterRoleBindingName, ns: "", group: "rbac.authorization.k8s.io", version: "v1", kind: "ClusterRoleBinding"},
			{name: render.GuardianDeploymentName, ns: render.GuardianNamespace, group: "apps", version: "v1", kind: "Deployment"},
			{name: render.GuardianConfigMapName, ns: render.GuardianNamespace, group: "", version: "v1", kind: "ConfigMap"},
		}
		i := 0
		for _, expectedRes := range expectedResources {
			ExpectResource(resources[i], expectedRes.name, expectedRes.ns, expectedRes.group, expectedRes.version, expectedRes.kind)
			i++
		}
	})

})
