/*
Copyright 2018 The Kubernetes Authors.

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

package v1alpha1

import (
	app "github.com/kubernetes-sigs/application/pkg/apis/app/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"log"
	crscheme "sigs.k8s.io/controller-runtime/pkg/runtime/scheme"
	"sigs.k8s.io/kubesdk/pkg/component"
	"sigs.k8s.io/kubesdk/pkg/resource"
	"sigs.k8s.io/kubesdk/pkg/resource/manager"
	"sigs.k8s.io/kubesdk/pkg/resource/manager/k8s"
	"sigs.k8s.io/kubesdk/pkg/status"
)

var (
	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &crscheme.Builder{
		GroupVersion: schema.GroupVersion{
			Group:   "foo.cloud.google.com",
			Version: "v1alpha1",
		},
	}
)

// Foo test custom resource
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Foo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FooSpec   `json:"spec,omitempty"`
	Status            FooStatus `json:"status,omitempty"`
}

// FooSpec CR foo spec
type FooSpec struct {
	Version string
}

// FooList contains a list of Foo
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type FooList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Foo `json:"items"`
}

// FooStatus CR foo status
type FooStatus struct {
	Status    string
	Component string
}

// ExpectedResources - returns resources
func (s *FooSpec) ExpectedResources(rsrc interface{}, rsrclabels map[string]string, dependent, aggregate *resource.Bag) (*resource.Bag, error) {
	var resources *resource.Bag = new(resource.Bag)
	r := rsrc.(*Foo)
	n := r.ObjectMeta.Name
	ns := r.ObjectMeta.Namespace
	resources.Add(
		[]resource.Item{
			{
				Lifecycle: resource.LifecycleManaged,
				Type:      k8s.Type,
				Obj: &k8s.Object{
					ObjList: &appsv1.DeploymentList{},
					Obj: &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      n + "-deploy",
							Namespace: ns,
							Labels:    rsrclabels,
						},
					},
				},
			},
			{
				Lifecycle: resource.LifecycleManaged,
				Type:      k8s.Type,
				Obj: &k8s.Object{
					ObjList: &corev1.ConfigMapList{},
					Obj: &corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      n + "-cm",
							Namespace: ns,
							Labels:    rsrclabels,
						},
						Data: map[string]string{
							"test-key": "test-value",
						},
					},
				},
			},
		}...,
	)
	return resources, nil
}

// Observables - return selectors
func (s *FooSpec) Observables(rsrcmgr *manager.ResourceManager, rsrc interface{}, rsrclabels map[string]string, expected *resource.Bag) []resource.Observable {
	return []resource.Observable{
		{
			Type: k8s.Type,
			Obj: &k8s.Observable{
				ObjList: &appsv1.DeploymentList{},
				Labels:  rsrclabels,
				Type:    metav1.TypeMeta{Kind: "Deployment", APIVersion: "apps/v1"},
			},
		},
		{
			Type: k8s.Type,
			Obj: &k8s.Observable{
				ObjList: &corev1.ConfigMapList{},
				Labels:  rsrclabels,
				Type:    metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"},
			},
		},
	}
}

// UpdateComponentStatus - update status block
func (s *FooSpec) UpdateComponentStatus(rsrci interface{}, reconciled *resource.Bag, err error) {
	rsrc := rsrci.(*Foo)
	rsrcstatus := &rsrc.Status
	rsrcstatus.Component = "base " + status.StatusReady
}

// Update - update status block for ESStatus
func (s *FooStatus) Update(rsrc *Foo, reconciled []metav1.Object) {
	s.Status = status.StatusReady
}

// ApplyDefaults applies defaults to the resource
func (r *Foo) ApplyDefaults() {
	if r.Spec.Version == "" {
		r.Spec.Version = "v1.0"
	}
}

// Validate validates the spec
func (r *Foo) Validate() error {
	return nil
}

// HandleError records status or error in status
func (r *Foo) HandleError(err error) {
}

// Application return app obj
func (s *FooSpec) Application(rsrc interface{}) app.Application {
	return app.Application{}
}

// Components returns components for this resource
func (r *Foo) Components() []component.Component {
	return []component.Component{
		{
			Handle:   &r.Spec,
			Name:     "escluster",
			CR:       r,
			OwnerRef: r.OwnerRef(),
		},
	}
}

// OwnerRef returns owner ref object with the component's resource as owner
func (r *Foo) OwnerRef() *metav1.OwnerReference {
	return metav1.NewControllerRef(r, schema.GroupVersionKind{
		Group:   "foobar",
		Version: "v1alpha1",
		Kind:    "Foo",
	})
}

func init() {
	SchemeBuilder.Register(&Foo{}, &FooList{})
	err := SchemeBuilder.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatal(err)
	}
}
