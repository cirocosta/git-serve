package v1alpha1

import (
	"fmt"
	"net/url"

	"github.com/vmware-labs/reconciler-runtime/apis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GitServerSpec struct {
	// Image is the image to use for the deployment of gitserver.
	//
	Image string `json:"image"`

	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
}

// GitServerStatus defines the observed state of GitServer
//
type GitServerStatus struct {
	apis.Status `json:",inline"`

	DeploymentRef *TypedLocalObjectReference `json:"deploymentRef,omitempty"`
	ServiceRef    *TypedLocalObjectReference `json:"serviceRef,omitempty"`
	Address       *Addressable               `json:"address,omitempty"`
}

// +k8s:deepcopy-gen=true
type Addressable struct {
	URL string `json:"url,omitempty"`
}

func (a *Addressable) Parse() (*url.URL, error) {
	return url.Parse(a.URL)
}

type TypedLocalObjectReference struct {
	// APIGroup is the group for the resource being referenced.
	// If APIGroup is not specified, the specified Kind must be in the core API group.
	// For any other third-party types, APIGroup is required.
	// +optional
	// +nullable
	APIGroup *string `json:"apiGroup" protobuf:"bytes,1,opt,name=apiGroup"`
	// Kind is the type of resource being referenced
	Kind string `json:"kind" protobuf:"bytes,2,opt,name=kind"`
	// Name is the name of resource being referenced
	Name string `json:"name" protobuf:"bytes,3,opt,name=name"`
}

func NewTypedLocalObjectReference(name string, gk schema.GroupKind) *TypedLocalObjectReference {
	if name == "" || gk.Empty() {
		return nil
	}

	ref := &TypedLocalObjectReference{
		Kind: gk.Kind,
		Name: name,
	}
	if gk.Group != "" && gk.Group != "core" {
		ref.APIGroup = &gk.Group
	}
	return ref
}

func NewTypedLocalObjectReferenceForObject(obj client.Object, scheme *runtime.Scheme) *TypedLocalObjectReference {
	if obj == nil {
		return nil
	}

	gvks, _, err := scheme.ObjectKinds(obj)
	if err != nil || len(gvks) == 0 {
		panic(fmt.Errorf("Unregistered runtime object: %v", err))
	}
	return NewTypedLocalObjectReference(obj.GetName(), gvks[0].GroupKind())
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="URL",type=string,JSONPath=`.status.address.url`
//+kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
//+kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].reason`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

type GitServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitServerSpec   `json:"spec,omitempty"`
	Status GitServerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type GitServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GitServer{}, &GitServerList{})
}
