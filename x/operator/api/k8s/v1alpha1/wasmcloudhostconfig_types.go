/*
Copyright 2024.

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
	"go.wasmcloud.dev/operator/api/condition"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PolicyService struct {
	//+kubebuilder:validation:Optional
	Topic string `json:"topic,omitempty"`
	//+kubebuilder:validation:Optional
	TimeoutMs int32 `json:"timeoutMs,omitempty"`
	//+kubebuilder:validation:Optional
	ChangesTopic string `json:"changesTopic,omitempty"`
}

type WasmCloudHostConfigResources struct {
	//+kubebuilder:validation:Optional
	Nats *corev1.ResourceRequirements `json:"nats,omitempty"`
	//+kubebuilder:validation:Optional
	Wasmcloud *corev1.ResourceRequirements `json:"wasmcloud,omitempty"`
}

type KubernetesSchedulingOptions struct {
	//+kubebuilder:validation:Optional
	DaemonSet bool `json:"daemonset,omitempty"`
	//+kubebuilder:validation:Schemaless
	//+kubebuilder:validation:Type=object
	PodTemplateAdditions *corev1.PodSpec `json:"podTemplateAdditions,omitempty"`
	//+kubebuilder:validation:Optional
	Resources *WasmCloudHostConfigResources `json:"resources,omitempty"`
}

type OtelSignalConfiguration struct {
	//+kubebuilder:validation:Optional
	Enabled bool `json:"enabled,omitempty"`
	//+kubebuilder:validation:Optional
	Endpoint string `json:"endpoint,omitempty"`
}

type ObservabilityConfiguration struct {
	//+kubebuilder:validation:Optional
	Enabled bool `json:"enabled,omitempty"`
	//+kubebuilder:validation:Optional
	Endpoint string `json:"endpoint,omitempty"`
	//+kubebuilder:validation:Optional
	//+kubebuilder:validation:Enum=grpc;http
	Protocol string `json:"protocol,omitempty"`
	//+kubebuilder:validation:Optional
	Logs *OtelSignalConfiguration `json:"logs,omitempty"`
	//+kubebuilder:validation:Optional
	Metrics *OtelSignalConfiguration `json:"metrics,omitempty"`
	//+kubebuilder:validation:Optional
	Traces *OtelSignalConfiguration `json:"traces,omitempty"`
}

type WasmCloudHostCertificates struct {
	//+kubebuilder:validation:Optional
	Authorities []corev1.Volume `json:"authorities,omitempty"`
}

// WasmCloudHostConfigSpec defines the desired state of WasmCloudHostConfig.
type WasmCloudHostConfigSpec struct {
	//+kubebuilder:validation:Required
	Version string `json:"version"`
	//+kubebuilder:validation:Required
	//+kubebuilder:default="default"
	Lattice string `json:"lattice,omitempty"`

	//+kubebuilder:validation:Optional
	HostReplicas *int32 `json:"hostReplicas,omitempty"`
	//+kubebuilder:validation:Optional
	HostLabels map[string]string `json:"hostLabels,omitempty"`
	//+kubebuilder:validation:Optional
	Image string `json:"image,omitempty"`
	//+kubebuilder:validation:Optional
	NatsLeafImage string `json:"natsLeafImage,omitempty"`
	//+kubebuilder:validation:Optional
	SecretName string `json:"secretName,omitempty"`
	//+kubebuilder:validation:Optional
	EnableStructuredLogging bool `json:"enableStructuredLogging,omitempty"`
	//+kubebuilder:validation:Optional
	RegistryCredentialsSecret string `json:"registryCredentialsSecret,omitempty"`
	//+kubebuilder:validation:Optional
	ControlServiceEnabled bool `json:"controlServiceEnabled,omitempty"`
	//+kubebuilder:validation:Optional
	ControlTopicPrefix string `json:"controlTopicPrefix,omitempty"`
	//+kubebuilder:validation:Optional
	//+kubebuilder:default="leaf"
	LeafNodeDomain string `json:"leafNodeDomain,omitempty"`
	//+kubebuilder:validation:Optional
	//+kubebuilder:default="nats://nats.default.svc.cluster.local"
	NatsAddress string `json:"natsAddress,omitempty"`
	//+kubebuilder:validation:Optional
	//+kubebuilder:default=4222
	NatsClientPort int `json:"natsClientPort,omitempty"`
	//+kubebuilder:validation:Optional
	//+kubebuilder:default=7422
	NatsLeafnodePort int `json:"natsLeafnodePort,omitempty"`
	//+kubebuilder:validation:Optional
	//+kubebuilder:default="default"
	JetstreamDomain string `json:"jetstreamDomain,omitempty"`
	//+kubebuilder:validation:Optional
	AllowLatest bool `json:"allowLatest,omitempty"`
	//+kubebuilder:validation:Optional
	AllowInsecure []string `json:"allowInsecure,omitempty"`
	//+kubebuilder:validation:Optional
	//+kubebuilder:default="INFO"
	LogLevel string `json:"logLevel,omitempty"`
	//+kubebuilder:validation:Optional
	PolicyService *PolicyService `json:"policyService,omitempty"`
	//+kubebuilder:validation:Optional
	SchedulingOptions *KubernetesSchedulingOptions `json:"schedulingOptions,omitempty"`
	//+kubebuilder:validation:Optional
	Observability *ObservabilityConfiguration `json:"observability,omitempty"`
	//+kubebuilder:validation:Optional
	Certificates *WasmCloudHostCertificates `json:"certificates,omitempty"`
	//+kubebuilder:validation:Optional
	SecretsTopicPrefix string `json:"secretsTopicPrefix,omitempty"`
	//+kubebuilder:validation:Optional
	MaxLinearMemoryBytes *uint32 `json:"maxLinearMemoryBytes,omitempty"`
}

// WasmCloudHostConfigStatus defines the observed state of WasmCloudHostConfig.
type WasmCloudHostConfigStatus struct {
	condition.ConditionedStatus `json:",inline"`

	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	//+kubebuilder:validation:Optional
	Apps []ApplicationStatus `json:"apps"`
	// NOTE(lxf): This should be 'appCount', 'app_count' is coming from the Rust Operator.
	//+kubebuilder:validation:Optional
	AppCount uint32 `json:"app_count"`
}

type ApplicationStatus struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	// NOTE(lxf): This should also expose the status of the application
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories={wasmcloud},shortName={whc}
// +kubebuilder:printcolumn:name="APPCOUNT",type=integer,JSONPath=`.status.appCount`
// +kubebuilder:printcolumn:name="AGE",type=date,JSONPath=".metadata.creationTimestamp"
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WasmCloudHostConfig is the Schema for the wasmcloudhostconfigs API.
type WasmCloudHostConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	//+kubebuilder:validation:Required
	Spec WasmCloudHostConfigSpec `json:"spec,omitempty"`

	//+kubebuilder:validation:Optional
	Status WasmCloudHostConfigStatus `json:"status,omitempty"`
}

func (w *WasmCloudHostConfig) SetCondition(c ...condition.Condition) {
	w.Status.SetConditions(c...)
}

func (w *WasmCloudHostConfig) GetCondition(ct condition.ConditionType) condition.Condition {
	return w.Status.GetCondition(ct)
}

// +kubebuilder:object:root=true

// WasmCloudHostConfigList contains a list of WasmCloudHostConfig.
type WasmCloudHostConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WasmCloudHostConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WasmCloudHostConfig{}, &WasmCloudHostConfigList{})
}
