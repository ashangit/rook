/*
Copyright 2018 The Rook Authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package v1alpha2

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (p PlacementSpec) All() Placement {
	return p[KeyAll]
}

/*
 * Pod spec
 */

func (p Placement) SetPodPlacement(pod *v1.PodSpec, nodeSelector map[string]string, isHostNetwork, allowMultiplePerNode bool, matchLabels map[string]string) {
	p.ApplyToPodSpec(pod)
	pod.NodeSelector = nodeSelector

	// when a node selector is being used, skip the affinity business below
	if nodeSelector != nil {
		return
	}

	// label selector for monitors used in anti-affinity rules
	podAntiAffinity := v1.PodAffinityTerm{
		LabelSelector: &metav1.LabelSelector{
			MatchLabels: matchLabels,
		},
		TopologyKey: v1.LabelHostname,
	}

	// set monitor pod anti-affinity rules. when monitors should never be
	// co-located (e.g. not AllowMultiplePerHost or HostNetworking) then the
	// anti-affinity rule is made to be required during scheduling, otherwise it
	// is merely a preferred policy.
	//
	// ApplyToPodSpec ensures that pod.Affinity is non-nil
	if pod.Affinity.PodAntiAffinity == nil {
		pod.Affinity.PodAntiAffinity = &v1.PodAntiAffinity{}
	}
	paa := pod.Affinity.PodAntiAffinity

	if isHostNetwork || !allowMultiplePerNode {
		paa.RequiredDuringSchedulingIgnoredDuringExecution =
			append(paa.RequiredDuringSchedulingIgnoredDuringExecution, podAntiAffinity)
	} else {
		paa.PreferredDuringSchedulingIgnoredDuringExecution =
			append(paa.PreferredDuringSchedulingIgnoredDuringExecution, v1.WeightedPodAffinityTerm{
				Weight:          50,
				PodAffinityTerm: podAntiAffinity,
			})
	}
}

// ApplyToPodSpec adds placement to a pod spec
func (p Placement) ApplyToPodSpec(t *v1.PodSpec) {
	if t.Affinity == nil {
		t.Affinity = &v1.Affinity{}
	}
	if p.NodeAffinity != nil {
		t.Affinity.NodeAffinity = p.NodeAffinity
	}
	if p.PodAffinity != nil {
		t.Affinity.PodAffinity = p.PodAffinity
	}
	if p.PodAntiAffinity != nil {
		t.Affinity.PodAntiAffinity = p.PodAntiAffinity
	}

	if p.Tolerations != nil {
		t.Tolerations = p.Tolerations
	}
}

// Merge returns a Placement which results from merging the attributes of the
// original Placement with the attributes of the supplied one. The supplied
// Placement's attributes will override the original ones if defined.
func (p Placement) Merge(with Placement) Placement {
	ret := p
	if with.NodeAffinity != nil {
		ret.NodeAffinity = with.NodeAffinity
	}
	if with.PodAffinity != nil {
		ret.PodAffinity = with.PodAffinity
	}
	if with.PodAntiAffinity != nil {
		ret.PodAntiAffinity = with.PodAntiAffinity
	}
	if with.Tolerations != nil {
		ret.Tolerations = with.Tolerations
	}
	return ret
}
