// Copyright (c) 2019 Tigera, Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package render_test

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/library-go/pkg/crypto"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	operator "github.com/tigera/operator/pkg/apis/operator/v1"
	"github.com/tigera/operator/pkg/render"
)

var _ = Describe("Tigera Secure Manager rendering tests", func() {
	var instance *operator.Manager
	oidcEnvVar := corev1.EnvVar{
		Name:      "CNX_WEB_OIDC_AUTHORITY",
		Value:     "",
		ValueFrom: nil,
	}
	BeforeEach(func() {
		// Initialize a default instance to use. Each test can override this to its
		// desired configuration.
		instance = &operator.Manager{
			Spec: operator.ManagerSpec{
				Auth: &operator.Auth{
					Type: operator.AuthTypeBasic,
				},
			},
		}
	})

	It("should render all resources for a default configuration", func() {
		resources := renderObjects(instance, nil)
		Expect(len(resources)).To(Equal(22))

		// Should render the correct resources.
		expectedResources := []struct {
			name    string
			ns      string
			group   string
			version string
			kind    string
		}{
			{name: "tigera-manager", ns: "", group: "", version: "v1", kind: "Namespace"},
			{name: "tigera-manager", ns: "tigera-manager", group: "", version: "v1", kind: "ServiceAccount"},
			{name: "tigera-manager-role", ns: "", group: "rbac.authorization.k8s.io", version: "v1", kind: "ClusterRole"},
			{name: "tigera-manager-binding", ns: "", group: "rbac.authorization.k8s.io", version: "v1", kind: "ClusterRoleBinding"},
			{name: "tigera-manager-pip", ns: "", group: "rbac.authorization.k8s.io", version: "v1", kind: "ClusterRole"},
			{name: "tigera-manager-pip", ns: "", group: "rbac.authorization.k8s.io", version: "v1", kind: "ClusterRoleBinding"},
			{name: "manager-tls", ns: "tigera-operator", group: "", version: "v1", kind: "Secret"},
			{name: "manager-tls", ns: "tigera-manager", group: "", version: "v1", kind: "Secret"},
			{name: "tigera-manager", ns: "tigera-manager", group: "", version: "v1", kind: "Service"},
			{name: "tigera-ui-user", ns: "", group: "rbac.authorization.k8s.io", version: "v1", kind: "ClusterRole"},
			{name: "tigera-network-admin", ns: "", group: "rbac.authorization.k8s.io", version: "v1", kind: "ClusterRole"},
			{name: render.VoltronTunnelSecretName, ns: "tigera-operator", group: "", version: "v1", kind: "Secret"},
			{name: render.VoltronTunnelSecretName, ns: "tigera-manager", group: "", version: "v1", kind: "Secret"},
			{name: "tigera-manager", ns: "tigera-manager", group: "", version: "v1", kind: "Deployment"},
		}

		i := 0
		for _, expectedRes := range expectedResources {
			ExpectResource(resources[i], expectedRes.name, expectedRes.ns, expectedRes.group, expectedRes.version, expectedRes.kind)
			i++
		}
	})

	It("should ensure cnx policy recommendation support is always set to true", func() {
		resources := renderObjects(instance, nil)
		Expect(len(resources)).To(Equal(22))

		// Should render the correct resource based on test case.
		Expect(GetResource(resources, "tigera-manager", "tigera-manager", "", "v1", "Deployment")).ToNot(BeNil())

		d := resources[13].(*v1.Deployment)

		Expect(len(d.Spec.Template.Spec.Containers)).To(Equal(3))
		Expect(d.Spec.Template.Spec.Containers[0].Name).To(Equal("tigera-manager"))
		Expect(d.Spec.Template.Spec.Containers[0].Env[8].Name).To(Equal("CNX_POLICY_RECOMMENDATION_SUPPORT"))
		Expect(d.Spec.Template.Spec.Containers[0].Env[8].Value).To(Equal("true"))
	})

	It("should render OIDC configmaps given OIDC configuration", func() {
		instance.Spec.Auth.Type = operator.AuthTypeOIDC
		oidcConfig := &corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      render.ManagerOIDCConfig,
				Namespace: render.OperatorNamespace(),
			},
		}
		// Should render the correct resource based on test case.
		resources := renderObjects(instance, oidcConfig)
		Expect(len(resources)).To(Equal(23))

		Expect(GetResource(resources, render.ManagerOIDCConfig, "tigera-manager", "", "v1", "ConfigMap")).ToNot(BeNil())
		d := resources[14].(*v1.Deployment)

		Expect(d.Spec.Template.Spec.Containers[0].Env).To(ContainElement(oidcEnvVar))

		// Make sure well-known and JWKS are accessible from manager.
		Expect(len(d.Spec.Template.Spec.Containers[0].VolumeMounts)).To(Equal(3))
		Expect(d.Spec.Template.Spec.Containers[0].VolumeMounts[0].Name).To(Equal(render.ManagerOIDCConfig))
		Expect(d.Spec.Template.Spec.Containers[0].VolumeMounts[0].MountPath).To(Equal(render.ManagerOIDCWellknownURI))
		Expect(d.Spec.Template.Spec.Containers[0].VolumeMounts[1].Name).To(Equal(render.ManagerOIDCConfig))
		Expect(d.Spec.Template.Spec.Containers[0].VolumeMounts[1].MountPath).To(Equal(render.ManagerOIDCJwksURI))

		Expect(len(d.Spec.Template.Spec.Volumes)).To(Equal(5))
		Expect(d.Spec.Template.Spec.Volumes[3].Name).To(Equal(render.ManagerOIDCConfig))
		Expect(d.Spec.Template.Spec.Volumes[3].ConfigMap.Name).To(Equal(render.ManagerOIDCConfig))
	})

	It("should set OIDC Authority environment when auth-type is OIDC", func() {
		instance.Spec.Auth.Type = operator.AuthTypeOIDC

		const authority = "https://foo.bar"
		instance.Spec.Auth.Authority = authority
		oidcEnvVar.Value = authority

		// Should render the correct resource based on test case.
		resources := renderObjects(instance, nil)
		Expect(len(resources)).To(Equal(22))
		d := resources[13].(*v1.Deployment)
		// tigera-manager volumes/volumeMounts checks.
		Expect(len(d.Spec.Template.Spec.Volumes)).To(Equal(4))
		Expect(d.Spec.Template.Spec.Containers[0].Env).To(ContainElement(oidcEnvVar))
		Expect(len(d.Spec.Template.Spec.Containers[0].VolumeMounts)).To(Equal(1))
	})

	It("should render multicluster settings properly", func() {
		resources := renderObjects(instance, nil)
		Expect(len(resources)).To(Equal(22))

		By("creating a valid self-signed cert")
		// Use the x509 package to validate that the cert was signed with the privatekey
		validateSecret(resources[11].(*corev1.Secret))
		validateSecret(resources[12].(*corev1.Secret))

		By("configuring the manager deployment")
		manager := resources[13].(*v1.Deployment).Spec.Template.Spec.Containers[0]
		Expect(manager.Name).To(Equal("tigera-manager"))
		ExpectEnv(manager.Env, "ENABLE_MULTI_CLUSTER_MANAGEMENT", "true")

		voltron := resources[13].(*v1.Deployment).Spec.Template.Spec.Containers[2]
		Expect(voltron.Name).To(Equal("tigera-voltron"))
		ExpectEnv(voltron.Env, "VOLTRON_ENABLE_MULTI_CLUSTER_MANAGEMENT", "true")
	})
})

func validateSecret(voltronSecret *corev1.Secret) {
	var newCert *x509.Certificate

	cert := voltronSecret.Data["cert"]
	key := voltronSecret.Data["key"]
	_, err := tls.X509KeyPair(cert, key)
	Expect(err).ShouldNot(HaveOccurred())

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(cert))
	Expect(ok).To(BeTrue())

	block, _ := pem.Decode([]byte(cert))
	Expect(err).ShouldNot(HaveOccurred())
	Expect(block).To(Not(BeNil()))

	newCert, err = x509.ParseCertificate(block.Bytes)
	Expect(err).ShouldNot(HaveOccurred())

	opts := x509.VerifyOptions{
		DNSName: render.VoltronDnsName,
		Roots:   roots,
	}

	_, err = newCert.Verify(opts)
	Expect(err).ShouldNot(HaveOccurred())

	opts = x509.VerifyOptions{
		DNSName:     render.VoltronDnsName,
		Roots:       x509.NewCertPool(),
		CurrentTime: time.Now().AddDate(0, 0, crypto.DefaultCACertificateLifetimeInDays+1),
	}
	_, err = newCert.Verify(opts)
	Expect(err).Should(HaveOccurred())

}

func renderObjects(instance *operator.Manager, oidcConfig *corev1.ConfigMap) []runtime.Object {
	esConfigMap := render.NewElasticsearchClusterConfig("clusterTestName", 1, 1)
	component, err := render.Manager(instance,
		nil,
		nil,
		esConfigMap,
		nil,
		nil,
		false,
		"",
		oidcConfig,
		true,
		nil)
	Expect(err).To(BeNil(), "Expected Manager to create successfully %s", err)
	resources := component.Objects()
	return resources
}
