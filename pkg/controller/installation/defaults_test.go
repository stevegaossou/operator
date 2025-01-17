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

package installation

import (
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	osconfigv1 "github.com/openshift/api/config/v1"
	operator "github.com/tigera/operator/pkg/apis/operator/v1"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("Defaulting logic tests", func() {
	It("should properly fill defaults on an empty instance", func() {
		instance := &operator.Installation{}
		fillDefaults(instance)
		Expect(instance.Spec.Variant).To(Equal(operator.Calico))
		Expect(instance.Spec.Registry).To(BeEmpty())
		Expect(instance.Spec.CalicoNetwork.IPPools).To(HaveLen(1))
		Expect(instance.Spec.CalicoNetwork.IPPools[0].CIDR).To(Equal("192.168.0.0/16"))
	})

	It("should properly fill defaults on an empty TigeraSecureEnterprise instance", func() {
		instance := &operator.Installation{}
		instance.Spec.Variant = operator.TigeraSecureEnterprise
		fillDefaults(instance)
		Expect(instance.Spec.Variant).To(Equal(operator.TigeraSecureEnterprise))
		Expect(instance.Spec.Registry).To(BeEmpty())
		Expect(instance.Spec.CalicoNetwork.IPPools).To(HaveLen(1))
		Expect(instance.Spec.CalicoNetwork.IPPools[0].CIDR).To(Equal("192.168.0.0/16"))
	})

	It("should error if CalicoNetwork is provided on EKS", func() {
		instance := &operator.Installation{}
		instance.Spec.Variant = operator.TigeraSecureEnterprise
		instance.Spec.CalicoNetwork = &operator.CalicoNetworkSpec{}
		instance.Spec.KubernetesProvider = operator.ProviderEKS
		Expect(fillDefaults(instance)).To(HaveOccurred())
	})

	It("should not override custom configuration", func() {
		var mtu int32 = 1500
		var ff bool = true
		instance := &operator.Installation{
			Spec: operator.InstallationSpec{
				Variant:  operator.TigeraSecureEnterprise,
				Registry: "test-reg/",
				ImagePullSecrets: []v1.LocalObjectReference{
					{
						Name: "pullSecret1",
					},
					{
						Name: "pullSecret2",
					},
				},
				CalicoNetwork: &operator.CalicoNetworkSpec{
					IPPools: []operator.IPPool{{
						CIDR:          "1.2.3.0/24",
						Encapsulation: "IPIPCrossSubnet",
						NATOutgoing:   "Enabled",
						NodeSelector:  "has(thiskey)",
					}},
					MTU: &mtu,
					NodeAddressAutodetectionV4: &operator.NodeAddressAutodetection{
						FirstFound: &ff,
					},
					NodeAddressAutodetectionV6: nil,
				},
			},
		}
		instanceCopy := instance.DeepCopyObject().(*operator.Installation)
		fillDefaults(instanceCopy)
		Expect(instanceCopy.Spec).To(Equal(instance.Spec))
	})

	It("should correct missing slashes on registry", func() {
		instance := &operator.Installation{
			Spec: operator.InstallationSpec{
				Registry: "test-reg",
			},
		}
		fillDefaults(instance)
		Expect(instance.Spec.Registry).To(Equal("test-reg/"))
	})

	table.DescribeTable("All pools should have all fields set from mergeAndFillDefaults function",
		func(i *operator.Installation, on *osconfigv1.Network) {
			Expect(mergeAndFillDefaults(i, on)).To(BeNil())
			if i.Spec.CalicoNetwork != nil && i.Spec.CalicoNetwork.IPPools != nil && len(i.Spec.CalicoNetwork.IPPools) == 1 {
				pool := i.Spec.CalicoNetwork.IPPools[0]
				Expect(pool.CIDR).ToNot(BeEmpty(), "CIDR should be set on pool %v", pool)
				Expect(pool.Encapsulation).To(BeElementOf(operator.EncapsulationTypes), "Encapsulation should be set on pool %q", pool)
				Expect(pool.NATOutgoing).To(BeElementOf(operator.NATOutgoingTypes), "NATOutgoing should be set on pool %v", pool)
				Expect(pool.NodeSelector).ToNot(BeEmpty(), "NodeSelector should be set on pool %v", pool)
			}
		},

		table.Entry("Empty config defaults IPPool", &operator.Installation{}, &osconfigv1.Network{}),
		table.Entry("Openshift only CIDR",
			&operator.Installation{
				Spec: operator.InstallationSpec{
					CalicoNetwork: &operator.CalicoNetworkSpec{},
				},
			}, &osconfigv1.Network{
				Spec: osconfigv1.NetworkSpec{
					ClusterNetwork: []osconfigv1.ClusterNetworkEntry{
						{CIDR: "10.0.0.0/8"},
					},
				},
			}),
		table.Entry("CIDR specified from OS config and Calico config",
			&operator.Installation{
				Spec: operator.InstallationSpec{
					CalicoNetwork: &operator.CalicoNetworkSpec{
						IPPools: []operator.IPPool{
							operator.IPPool{
								CIDR: "10.0.0.0/24",
							},
						},
					},
				},
			}, &osconfigv1.Network{
				Spec: osconfigv1.NetworkSpec{
					ClusterNetwork: []osconfigv1.ClusterNetworkEntry{
						{CIDR: "10.0.0.0/8"},
					},
				},
			}),
	)
})
