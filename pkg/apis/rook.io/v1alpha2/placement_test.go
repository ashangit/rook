/*
Copyright 2016 The Rook Authors. All rights reserved.

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
	"encoding/json"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestPlacement_spec(t *testing.T) {
	specYaml := []byte(`
nodeAffinity:
  requiredDuringSchedulingIgnoredDuringExecution:
    nodeSelectorTerms:
    - matchExpressions:
      - key: foo
        operator: In
        values:
          - bar
tolerations:
  - key: foo
    operator: Exists`)

	// convert the raw spec yaml into JSON
	rawJSON, err := yaml.YAMLToJSON(specYaml)
	assert.Nil(t, err)

	// unmarshal the JSON into a strongly typed placement spec object
	var placement Placement
	err = json.Unmarshal(rawJSON, &placement)
	assert.Nil(t, err)

	// the unmarshalled placement spec should equal the expected spec below
	expected := Placement{
		NodeAffinity: &v1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
				NodeSelectorTerms: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{
							{
								Key:      "foo",
								Operator: v1.NodeSelectorOpIn,
								Values:   []string{"bar"},
							},
						},
					},
				},
			},
		},
		Tolerations: []v1.Toleration{
			{
				Key:      "foo",
				Operator: v1.TolerationOpExists,
			},
		},
	}
	assert.Equal(t, expected, placement)
}

func TestPlacement_ApplyToPodSpec(t *testing.T) {
	to := placementTestGetTolerations("foo", "bar")
	na := placementTestGenerateNodeAffinity()
	expected := &v1.PodSpec{Affinity: &v1.Affinity{NodeAffinity: na}, Tolerations: to}

	var p Placement
	var ps *v1.PodSpec

	p = Placement{NodeAffinity: na, Tolerations: to}
	ps = &v1.PodSpec{}
	p.ApplyToPodSpec(ps)
	assert.Equal(t, expected, ps)

	// partial update
	p = Placement{NodeAffinity: na}
	ps = &v1.PodSpec{Tolerations: to}
	p.ApplyToPodSpec(ps)
	assert.Equal(t, expected, ps)

	// overridden attributes
	p = Placement{NodeAffinity: na, Tolerations: to}
	ps = &v1.PodSpec{Tolerations: placementTestGetTolerations("bar", "baz")}
	p.ApplyToPodSpec(ps)
	assert.Equal(t, expected, ps)

	p = Placement{NodeAffinity: na}
	nap := placementTestGenerateNodeAffinity()
	nap.PreferredDuringSchedulingIgnoredDuringExecution[0].Weight = 5
	ps = &v1.PodSpec{Affinity: &v1.Affinity{NodeAffinity: nap}, Tolerations: to}
	p.ApplyToPodSpec(ps)
	assert.Equal(t, expected, ps)
}

func TestPlacement_Merge(t *testing.T) {
	to := placementTestGetTolerations("foo", "bar")
	na := placementTestGenerateNodeAffinity()

	var original, with, expected, merged Placement

	original = Placement{}
	with = Placement{Tolerations: to}
	expected = Placement{Tolerations: to}
	merged = original.Merge(with)
	assert.Equal(t, expected, merged)

	original = Placement{NodeAffinity: na}
	with = Placement{Tolerations: to}
	expected = Placement{NodeAffinity: na, Tolerations: to}
	merged = original.Merge(with)
	assert.Equal(t, expected, merged)

	original = Placement{Tolerations: placementTestGetTolerations("bar", "baz")}
	with = Placement{NodeAffinity: na, Tolerations: to}
	expected = Placement{NodeAffinity: na, Tolerations: to}
	merged = original.Merge(with)
	assert.Equal(t, expected, merged)
}

func placementTestGetTolerations(key, value string) []v1.Toleration {
	var ts int64 = 10
	return []v1.Toleration{
		{
			Key:               key,
			Operator:          v1.TolerationOpExists,
			Value:             value,
			Effect:            v1.TaintEffectNoSchedule,
			TolerationSeconds: &ts,
		},
	}
}

func placementTestGenerateNodeAffinity() *v1.NodeAffinity {
	return &v1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
			NodeSelectorTerms: []v1.NodeSelectorTerm{
				{
					MatchExpressions: []v1.NodeSelectorRequirement{
						{
							Key:      "foo",
							Operator: v1.NodeSelectorOpExists,
							Values:   []string{"bar"},
						},
					},
				},
			},
		},
		PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
			{
				Weight: 10,
				Preference: v1.NodeSelectorTerm{
					MatchExpressions: []v1.NodeSelectorRequirement{
						{
							Key:      "foo",
							Operator: v1.NodeSelectorOpExists,
							Values:   []string{"bar"},
						},
					},
				},
			},
		},
	}
}

func testPodSpecPlacement(t *testing.T, hostNet, allowMulti bool, req, pref int, placement *Placement) {
	spec := v1.PodSpec{
		InitContainers: []v1.Container{},
		Containers:     []v1.Container{},
		RestartPolicy:  v1.RestartPolicyAlways,
		HostNetwork:    hostNet,
	}
	placement.SetPodPlacement(&spec, nil, hostNet, allowMulti, map[string]string{
		"app": "mon",
	})

	// should have a required anti-affinity and no preferred anti-affinity
	assert.Equal(t,
		req,
		len(spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution))
	assert.Equal(t,
		pref,
		len(spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution))
}

func makePlacement() Placement {
	return Placement{
		PodAntiAffinity: &v1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
				{
					TopologyKey: v1.LabelZoneFailureDomain,
				},
			},
			PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
				{
					PodAffinityTerm: v1.PodAffinityTerm{
						TopologyKey: v1.LabelZoneFailureDomain,
					},
				},
			},
		},
	}
}

func TestPodSpecPlacement(t *testing.T) {
	// no placement settings in the crd
	p := Placement{}
	testPodSpecPlacement(t, true, true, 1, 0, &p)
	testPodSpecPlacement(t, true, false, 1, 0, &p)
	testPodSpecPlacement(t, false, true, 0, 1, &p)
	testPodSpecPlacement(t, false, false, 1, 0, &p)

	// crd has other preferred and required anti-affinity setting
	p = makePlacement()
	testPodSpecPlacement(t, true, true, 2, 1, &p)
	p = makePlacement()
	testPodSpecPlacement(t, true, false, 2, 1, &p)
	p = makePlacement()
	testPodSpecPlacement(t, false, true, 1, 2, &p)
	p = makePlacement()
	testPodSpecPlacement(t, false, false, 2, 1, &p)
}
