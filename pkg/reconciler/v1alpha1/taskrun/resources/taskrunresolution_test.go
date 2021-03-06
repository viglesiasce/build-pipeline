/*
Copyright 2018 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either extress or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resources

import (
	"fmt"
	"testing"

	"github.com/knative/build-pipeline/pkg/apis/pipeline/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestResolveTaskRun(t *testing.T) {
	tr := &v1alpha1.TaskRunSpec{
		TaskRef: v1alpha1.TaskRef{
			Name: "orchestrate",
		},
		Inputs: v1alpha1.TaskRunInputs{
			Resources: []v1alpha1.TaskRunResource{{
				Name: "repoToBuildFrom",
				ResourceRef: v1alpha1.PipelineResourceRef{
					Name: "git-repo",
				},
			}, {
				Name: "clusterToUse",
				ResourceRef: v1alpha1.PipelineResourceRef{
					Name: "k8s-cluster",
				},
			}},
		},
		Outputs: v1alpha1.TaskRunOutputs{
			Resources: []v1alpha1.TaskRunResource{{
				Name: "imageToBuild",
				ResourceRef: v1alpha1.PipelineResourceRef{
					Name: "image",
				},
			}, {
				Name: "gitRepoToUpdate",
				ResourceRef: v1alpha1.PipelineResourceRef{
					Name: "another-git-repo",
				},
			}},
		},
	}

	task := &v1alpha1.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name: "orchestrate",
		},
	}
	gt := func(n string) (*v1alpha1.Task, error) { return task, nil }

	resources := []*v1alpha1.PipelineResource{{
		ObjectMeta: metav1.ObjectMeta{
			Name: "git-repo",
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{
			Name: "k8s-cluster",
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{
			Name: "image",
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{
			Name: "another-git-repo",
		},
	}}
	resourceIndex := 0
	gr := func(n string) (*v1alpha1.PipelineResource, error) {
		r := resources[resourceIndex]
		resourceIndex++
		return r, nil
	}

	rtr, err := ResolveTaskRun(tr, gt, gr)
	if err != nil {
		t.Fatalf("Did not expect error trying to resolve TaskRun: %s", err)
	}

	if rtr.Task == nil || rtr.Task.Name != "orchestrate" {
		t.Errorf("Task not resolved, expected `orchestrate` Task but got: %v", rtr.Task)
	}

	if len(rtr.Inputs) == 2 {
		r, ok := rtr.Inputs["repoToBuildFrom"]
		if !ok {
			t.Errorf("Expected value present in map for `repoToBuildFrom' but it was missing")
		} else {
			if r.Name != "git-repo" {
				t.Errorf("Expected to use resource `git-repo` for `repoToBuildFrom` but used %s", r.Name)
			}
		}
		r, ok = rtr.Inputs["clusterToUse"]
		if !ok {
			t.Errorf("Expected value present in map for `clusterToUse' but it was missing")
		} else {
			if r.Name != "k8s-cluster" {
				t.Errorf("Expected to use resource `k8s-cluster` for `clusterToUse` but used %s", r.Name)
			}
		}
	} else {
		t.Errorf("Expected 2 resolved inputs but instead had: %v", rtr.Inputs)
	}

	if len(rtr.Outputs) == 2 {
		r, ok := rtr.Outputs["imageToBuild"]
		if !ok {
			t.Errorf("Expected value present in map for `imageToBuild' but it was missing")
		} else {
			if r.Name != "image" {
				t.Errorf("Expected to use resource `image` for `imageToBuild` but used %s", r.Name)
			}
		}
		r, ok = rtr.Outputs["gitRepoToUpdate"]
		if !ok {
			t.Errorf("Expected value present in map for `gitRepoToUpdate' but it was missing")
		} else {
			if r.Name != "another-git-repo" {
				t.Errorf("Expected to use resource `another-git-repo` for `gitRepoToUpdate` but used %s", r.Name)
			}
		}
	} else {
		t.Errorf("Expected 2 resolved outputs but instead had: %v", rtr.Outputs)
	}
}
func TestResolveTaskRun_missingTask(t *testing.T) {
	tr := &v1alpha1.TaskRunSpec{
		TaskRef: v1alpha1.TaskRef{
			Name: "orchestrate",
		},
	}

	gt := func(n string) (*v1alpha1.Task, error) { return nil, fmt.Errorf("nope") }
	gr := func(n string) (*v1alpha1.PipelineResource, error) { return &v1alpha1.PipelineResource{}, nil }

	_, err := ResolveTaskRun(tr, gt, gr)
	if err == nil {
		t.Fatalf("Expected to get error because task couldn't be resolved")
	}
}
func TestResolveTaskRun_missingOutput(t *testing.T) {
	tr := &v1alpha1.TaskRunSpec{
		TaskRef: v1alpha1.TaskRef{
			Name: "orchestrate",
		},
		Outputs: v1alpha1.TaskRunOutputs{
			Resources: []v1alpha1.TaskRunResource{{
				Name: "repoToUpdate",
				ResourceRef: v1alpha1.PipelineResourceRef{
					Name: "another-git-repo",
				},
			},
			},
		}}

	gt := func(n string) (*v1alpha1.Task, error) { return &v1alpha1.Task{}, nil }
	gr := func(n string) (*v1alpha1.PipelineResource, error) { return nil, fmt.Errorf("nope") }

	_, err := ResolveTaskRun(tr, gt, gr)
	if err == nil {
		t.Fatalf("Expected to get error because output resource couldn't be resolved")
	}
}

func TestResolveTaskRun_missingInput(t *testing.T) {
	tr := &v1alpha1.TaskRunSpec{
		TaskRef: v1alpha1.TaskRef{
			Name: "orchestrate",
		},
		Inputs: v1alpha1.TaskRunInputs{
			Resources: []v1alpha1.TaskRunResource{{
				Name: "repoToBuildFrom",
				ResourceRef: v1alpha1.PipelineResourceRef{
					Name: "git-repo",
				},
			},
			},
		}}

	gt := func(n string) (*v1alpha1.Task, error) { return &v1alpha1.Task{}, nil }
	gr := func(n string) (*v1alpha1.PipelineResource, error) { return nil, fmt.Errorf("nope") }

	_, err := ResolveTaskRun(tr, gt, gr)
	if err == nil {
		t.Fatalf("Expected to get error because input resource couldn't be resolved")
	}
}

func TestResolveTaskRun_noResources(t *testing.T) {
	tr := &v1alpha1.TaskRunSpec{
		TaskRef: v1alpha1.TaskRef{
			Name: "orchestrate",
		},
	}
	task := &v1alpha1.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name: "orchestrate",
		},
	}
	gt := func(n string) (*v1alpha1.Task, error) { return task, nil }
	gr := func(n string) (*v1alpha1.PipelineResource, error) { return &v1alpha1.PipelineResource{}, nil }

	rtr, err := ResolveTaskRun(tr, gt, gr)
	if err != nil {
		t.Fatalf("Did not expect error trying to resolve TaskRun: %s", err)
	}

	if rtr.Task == nil || rtr.Task.Name != "orchestrate" {
		t.Errorf("Task not resolved, expected `orchestrate` Task but got: %v", rtr.Task)
	}

	if len(rtr.Inputs) != 0 {
		t.Errorf("Did not expect any outputs to be resolved when none specified but had %v", rtr.Inputs)
	}
	if len(rtr.Outputs) != 0 {
		t.Errorf("Did not expect any outputs to be resolved when none specified but had %v", rtr.Outputs)
	}
}
