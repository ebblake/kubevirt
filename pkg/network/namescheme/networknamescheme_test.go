/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2022 Red Hat, Inc.
 *
 */

package namescheme_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	networkv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"

	virtv1 "kubevirt.io/api/core/v1"

	"kubevirt.io/kubevirt/pkg/network/namescheme"
)

var _ = Describe("Network Name Scheme", func() {
	DescribeTable("create pod interfaces name scheme",
		func(nameSchemeFn func([]virtv1.Network) map[string]string, networkList []virtv1.Network, expectedNetworkNameScheme map[string]string) {
			podIfacesNameScheme := nameSchemeFn(networkList)

			Expect(podIfacesNameScheme).To(Equal(expectedNetworkNameScheme))
		},
		Entry("hashed, when network list is nil", namescheme.CreateHashedNetworkNameScheme, nil, map[string]string{}),
		Entry("hashed, when no multus networks exist",
			namescheme.CreateHashedNetworkNameScheme,
			[]virtv1.Network{
				newPodNetwork("default"),
			},
			map[string]string{
				"default": namescheme.PrimaryPodInterfaceName,
			}),
		Entry("hashed, when default multus networks exist",
			namescheme.CreateHashedNetworkNameScheme,
			[]virtv1.Network{
				createMultusDefaultNetwork("network0", "default/nad0"),
				createMultusSecondaryNetwork("network1", "default/nad1"),
				createMultusSecondaryNetwork("network2", "default/nad2"),
			},
			map[string]string{
				"network0": namescheme.PrimaryPodInterfaceName,
				"network1": "poda7662f44d65",
				"network2": "pod27f4a77f94e",
			}),
		Entry("hashed, when default pod networks exist",
			namescheme.CreateHashedNetworkNameScheme,
			[]virtv1.Network{
				newPodNetwork("default"),
				createMultusSecondaryNetwork("network1", "default/nad1"),
				createMultusSecondaryNetwork("network2", "default/nad2"),
			},
			map[string]string{
				"default":  namescheme.PrimaryPodInterfaceName,
				"network1": "poda7662f44d65",
				"network2": "pod27f4a77f94e",
			}),
		Entry("ordinal, when network list is nil", namescheme.CreateOrdinalNetworkNameScheme, nil, map[string]string{}),
		Entry("ordinal, when no multus networks exist",
			namescheme.CreateOrdinalNetworkNameScheme,
			[]virtv1.Network{
				newPodNetwork("default"),
			},
			map[string]string{
				"default": namescheme.PrimaryPodInterfaceName,
			}),
		Entry("ordinal, when default multus networks exist",
			namescheme.CreateOrdinalNetworkNameScheme,
			[]virtv1.Network{
				createMultusDefaultNetwork("network0", "default/nad0"),
				createMultusSecondaryNetwork("network1", "default/nad1"),
				createMultusSecondaryNetwork("network2", "default/nad2"),
			},
			map[string]string{
				"network0": namescheme.PrimaryPodInterfaceName,
				"network1": "net1",
				"network2": "net2",
			}),
		Entry("ordinal, when default pod networks exist",
			namescheme.CreateOrdinalNetworkNameScheme,
			[]virtv1.Network{
				newPodNetwork("default"),
				createMultusSecondaryNetwork("network1", "default/nad1"),
				createMultusSecondaryNetwork("network2", "default/nad2"),
			},
			map[string]string{
				"default":  namescheme.PrimaryPodInterfaceName,
				"network1": "net1",
				"network2": "net2",
			}),
	)

	Context("CreateNetworkNameSchemeByPodNetworkStatus", func() {
		const (
			redNetworkName      = "red"
			redIfaceHashedName  = "podb1f51a511f1"
			redIfaceOrdinalName = "net1"

			greenNetworkName      = "green"
			greenIfaceHashedName  = "podba4788b226a"
			greenIfaceOrdinalName = "net2"
		)

		DescribeTable("should return mapping between VMI networks and the pod interfaces names according to Multus network-status annotation",
			func(networks []virtv1.Network, networkStatus map[string]networkv1.NetworkStatus, expectedNameScheme map[string]string) {
				Expect(namescheme.CreateNetworkNameSchemeByPodNetworkStatus(networks, networkStatus)).To(Equal(expectedNameScheme))
			},
			Entry("when Multus network-status annotation is absent",
				multusNetworks(redNetworkName, greenNetworkName),
				map[string]networkv1.NetworkStatus{},
				map[string]string{
					redNetworkName:   redIfaceHashedName,
					greenNetworkName: greenIfaceHashedName,
				},
			),
			Entry("given only pod network",
				[]virtv1.Network{newPodNetwork("default")},
				map[string]networkv1.NetworkStatus{
					"default": {Interface: namescheme.PrimaryPodInterfaceName},
				},
				map[string]string{
					"default": namescheme.PrimaryPodInterfaceName,
				},
			),
			Entry("when the annotation reflects the pod interfaces naming is hashed",
				multusNetworks(redNetworkName, greenNetworkName),
				map[string]networkv1.NetworkStatus{
					redNetworkName:   {Interface: redIfaceHashedName},
					greenNetworkName: {Interface: greenIfaceHashedName},
				},
				map[string]string{
					redNetworkName:   redIfaceHashedName,
					greenNetworkName: greenIfaceHashedName,
				},
			),
			Entry("when the annotation reflects the pod interface naming is ordinal",
				multusNetworks(redNetworkName, greenNetworkName),
				map[string]networkv1.NetworkStatus{
					redNetworkName:   {Interface: redIfaceOrdinalName},
					greenNetworkName: {Interface: greenIfaceOrdinalName},
				},
				map[string]string{
					redNetworkName:   redIfaceOrdinalName,
					greenNetworkName: greenIfaceOrdinalName,
				},
			),
		)
	})
})

func multusNetworks(names ...string) []virtv1.Network {
	var networks []virtv1.Network
	for _, name := range names {
		networks = append(networks, createMultusNetwork(name, name+"net"))
	}
	return networks
}

func createMultusSecondaryNetwork(name, networkName string) virtv1.Network {
	return createMultusNetwork(name, networkName)
}

func createMultusDefaultNetwork(name, networkName string) virtv1.Network {
	multusNetwork := createMultusNetwork(name, networkName)
	multusNetwork.Multus.Default = true
	return multusNetwork
}

func createMultusNetwork(name, networkName string) virtv1.Network {
	return virtv1.Network{
		Name: name,
		NetworkSource: virtv1.NetworkSource{
			Multus: &virtv1.MultusNetwork{
				NetworkName: networkName,
			},
		},
	}
}

func newPodNetwork(name string) virtv1.Network {
	return virtv1.Network{
		Name: name,
		NetworkSource: virtv1.NetworkSource{
			Pod: &virtv1.PodNetwork{},
		},
	}
}
