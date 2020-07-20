package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WorkLoadPlanSpec defines the desired state of WorkLoadPlan
// +k8s:openapi-gen=true
type WorkLoadPlanSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	Disable bool `json:"disable"`
	Cron *CrondPlan `json:"cron,omitempty"`
	Disposable *DisposablePlan `json:"disposable,omitempty"`
	WorkLoads []*WorkLoad `json:"workloads,omitempty"`
}

// WorkLoadPlanStatus defines the observed state of WorkLoadPlan
// +k8s:openapi-gen=true
type WorkLoadPlanStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkLoadPlan is the Schema for the workloadplans API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=workloadplans,scope=Namespaced
type WorkLoadPlan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkLoadPlanSpec   `json:"spec,omitempty"`
	Status WorkLoadPlanStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkLoadPlanList contains a list of WorkLoadPlan
type WorkLoadPlanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkLoadPlan `json:"items"`
}

type CrondPlan struct {

	Period string `json:"period,omitempty"`		//year month day

	StartSecond string `json:"startSecond,omitempty"`
	EndSecond string `json:"endSecond,omitempty"`

	StartMinute string `json:"startMinute,omitempty"`
	EndMinute string `json:"endMinute,omitempty"`

	StartHour string `json:"startHour,omitempty"`
	EndHour string `json:"endHour,omitempty"`

	StartDay string `json:"startDay,omitempty"`
	EndDay string `json:"endDay,omitempty"`

	StartMonth string `json:"startMonth,omitempty"`
	EndMonth string `json:"endMonth,omitempty"`
}

type DisposablePlan struct {
	StartTime string `json:"startTime,omitempty"`
	EndTime string `json:"endTime,omitempty"`
}

type WorkLoad struct {
	WorkLoadType string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
}

func init() {
	SchemeBuilder.Register(&WorkLoadPlan{}, &WorkLoadPlanList{})
}
