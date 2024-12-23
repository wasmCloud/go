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

package k8s

import (
	"context"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"go.wasmcloud.dev/operator/api/condition"
	k8sv1alpha1 "go.wasmcloud.dev/operator/api/k8s/v1alpha1"
)

// WasmCloudHostConfigReconciler reconciles a WasmCloudHostConfig object
type WasmCloudHostConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=k8s.wasmcloud.dev,resources=wasmcloudhostconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=k8s.wasmcloud.dev,resources=wasmcloudhostconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=k8s.wasmcloud.dev,resources=wasmcloudhostconfigs/finalizers,verbs=update

// +kubebuilder:rbac:groups=apps,resources=deployments;daemonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments;daemonsets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments;daemonsets/finalizers,verbs=update

func (r *WasmCloudHostConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling WasmCloudHostConfig")

	var hostConfig k8sv1alpha1.WasmCloudHostConfig
	if err := r.Get(ctx, req.NamespacedName, &hostConfig); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !hostConfig.DeletionTimestamp.IsZero() {
		// The object is being deleted
		return ctrl.Result{}, nil
	}

	if hostConfig.Generation != hostConfig.Status.ObservedGeneration {
		hostConfig.SetCondition(condition.ReconcilePending())

		if err := r.reconcileWorkload(ctx, &hostConfig); err != nil {
			hostConfig.SetCondition(condition.ReconcileError(err))
		} else {
			hostConfig.Status.ObservedGeneration = hostConfig.Generation
			hostConfig.SetCondition(condition.ReconcileSuccess())
			return ctrl.Result{}, r.Status().Update(ctx, &hostConfig)
		}
	}

	if err := r.updateStatus(ctx, &hostConfig); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *WasmCloudHostConfigReconciler) updateStatus(ctx context.Context, hostConfig *k8sv1alpha1.WasmCloudHostConfig) error {
	// find deployment/daemonset
	// check if number of replicas is correct
	// flag as available
	// get app statuses
	hostConfig.SetCondition(condition.Available())
	return r.Status().Update(ctx, hostConfig)
}

func (r *WasmCloudHostConfigReconciler) reconcileWorkload(ctx context.Context, hostConfig *k8sv1alpha1.WasmCloudHostConfig) error {
	return r.reconcileDeployment(ctx, hostConfig)
}

func (r *WasmCloudHostConfigReconciler) reconcileDeployment(ctx context.Context, hostConfig *k8sv1alpha1.WasmCloudHostConfig) error {
	// host-label.k8s.wasmcloud.dev/<LABEL_NAME>: <LABEL_VALUE>

	wantLabels := map[string]string{
		"app.kubernetes.io/name":       "wasmcloud",
		"app.kubernetes.io/managed-by": "wasmcloud-operator",
		"app.kubernetes.io/instance":   hostConfig.GetName(),
	}

	defaultLabels := map[string]string{
		"app.kubernetes.io/name":       "wasmcloud",
		"app.kubernetes.io/managed-by": "wasmcloud-operator",
		"app.kubernetes.io/instance":   hostConfig.GetName(),
	}

	defaultEnv := []corev1.EnvVar{
		// k8s specific vars
		{
			Name: "WASMCLOUD_POD_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
		{
			Name: "WASMCLOUD_POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		{
			Name: "WASMCLOUD_POD_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		},
		{
			Name: "WASMCLOUD_NODE_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "spec.nodeName",
				},
			},
		},

		// wasmcloud specific vars
		{
			Name:  "WASMCLOUD_STRUCTURED_LOGGING_ENABLED",
			Value: strconv.FormatBool(hostConfig.Spec.EnableStructuredLogging),
		},
		{
			Name:  "WASMCLOUD_LOG_LEVEL",
			Value: hostConfig.Spec.LogLevel,
		},
		{
			Name:  "WASMCLOUD_JS_DOMAIN",
			Value: hostConfig.Spec.JetstreamDomain,
		},
		{
			Name:  "WASMCLOUD_LATTICE",
			Value: hostConfig.Spec.Lattice,
		},
		{
			Name:  "WASMCLOUD_NATS_HOST",
			Value: hostConfig.Spec.NatsAddress,
		},
		{
			Name:  "WASMCLOUD_NATS_PORT",
			Value: strconv.FormatInt(int64(hostConfig.Spec.NatsClientPort), 10),
		},
		{
			Name:  "WASMCLOUD_RPC_TIMEOUT_MS",
			Value: "4000",
		},
		{
			Name:  "WASMCLOUD_LABEL_kubernetes",
			Value: "true",
		},
	}

	for k, v := range hostConfig.Spec.HostLabels {
		defaultEnv = append(defaultEnv, corev1.EnvVar{
			Name:  "WASMCLOUD_LABEL_" + k,
			Value: v,
		})
	}

	volumes := []corev1.Volume{
		{
			Name: "wasmcloud-share",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}

	defaultMounts := []corev1.VolumeMount{
		{
			Name:      "wasmcloud-share",
			MountPath: "/share",
		},
	}

	hostImage := hostConfig.Spec.Image
	if hostImage == "" {
		hostImage = "ghcr.io/wasmcloud/wasmcloud"
	}
	hostImage = hostImage + ":" + hostConfig.Spec.Version

	host := corev1.Container{
		Name:  "wasmcloud-host",
		Image: hostImage,
		//EnvFrom:      mergeEnvFromSource(sandbox.Spec.EnvFrom),
		Env:          mergeEnvVar(defaultEnv),
		VolumeMounts: mergeMounts(defaultMounts),
		Ports: []corev1.ContainerPort{
			{
				Name:          "metrics",
				ContainerPort: 9090,
			},
		},
	}

	//	volumes = append(volumes, sandbox.Spec.Volumes...)

	podTemplate := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: mergeLabels(hostConfig.GetLabels(), defaultLabels),
		},
		Spec: corev1.PodSpec{
			EnableServiceLinks:            boolPtr(false),
			TerminationGracePeriodSeconds: int64Ptr(0),
			Containers:                    []corev1.Container{host},
			Volumes:                       volumes,
		},
	}

	spec := appsv1.DeploymentSpec{
		Replicas: hostConfig.Spec.HostReplicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: wantLabels,
		},
		Template: podTemplate,
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            hostConfig.GetName(),
			Namespace:       hostConfig.GetNamespace(),
			Labels:          mergeLabels(hostConfig.GetLabels(), defaultLabels),
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(hostConfig, hostConfig.GroupVersionKind())},
		},
		Spec: spec,
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, deployment, func() error {
		deployment.SetLabels(mergeLabels(deployment.GetLabels(), hostConfig.GetLabels(), wantLabels))

		// update spec but keep replicas stable
		// it might have been modified by hpa
		replicas := deployment.Spec.Replicas
		deployment.Spec = spec
		deployment.Spec.Replicas = replicas

		return controllerutil.SetControllerReference(hostConfig, deployment, r.Scheme)
	})

	return err
}

// SetupWithManager sets up the controller with the Manager.
func (r *WasmCloudHostConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&k8sv1alpha1.WasmCloudHostConfig{}).
		Owns(&appsv1.Deployment{}).
		Owns(&appsv1.DaemonSet{}).
		Named("k8s-wasmcloudhostconfig").
		Complete(r)
}
