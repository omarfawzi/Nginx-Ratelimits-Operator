package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// RateLimitsSpec defines the desired state of RateLimits
// +kubebuilder:object:generate=true
// +kubebuilder:resource:scope=Namespaced
// This CRD specifies a label selector for target pods
// and the rate limit configuration in YAML format.
type RateLimitsSpec struct {
	Selector metav1.LabelSelector `json:"selector"`
	// Env specifies the environment variables for the sidecar container.
	// Keys correspond to the variables documented in the README, e.g. UPSTREAM_HOST.
	Env map[string]string `json:"env,omitempty"`
	// RateLimits contains the rate limit configuration.
	// The structure follows the example in the README.
	RateLimits runtime.RawExtension `json:"rateLimits"`
}

// +kubebuilder:object:root=true
// RateLimits is the Schema for the ratelimitersidecars API
// This CR controls sidecar injection for matching pods.
type RateLimits struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec RateLimitsSpec `json:"spec,omitempty"`
}

func (in *RateLimits) DeepCopyInto(out *RateLimits) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
}

func (in *RateLimits) DeepCopy() *RateLimits {
	if in == nil {
		return nil
	}
	out := new(RateLimits)
	in.DeepCopyInto(out)
	return out
}

func (in *RateLimits) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}

// +kubebuilder:object:root=true
// RateLimitsList contains a list of RateLimits

type RateLimitsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RateLimits `json:"items"`
}

func (in *RateLimitsList) DeepCopyInto(out *RateLimitsList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]RateLimits, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *RateLimitsList) DeepCopy() *RateLimitsList {
	if in == nil {
		return nil
	}
	out := new(RateLimitsList)
	in.DeepCopyInto(out)
	return out
}

func (in *RateLimitsList) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}

// GroupVersion is group version used to register these objects
var GroupVersion = schema.GroupVersion{Group: "nginx.ratelimiter", Version: "v1alpha1"}

var SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

func init() {
	SchemeBuilder.Register(&RateLimits{}, &RateLimitsList{})
}

func AddToScheme(s *runtime.Scheme) error {
	return SchemeBuilder.AddToScheme(s)
}
