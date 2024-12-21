package k8s

import (
	"sort"

	corev1 "k8s.io/api/core/v1"
)

func mergeLabels(lbls ...map[string]string) map[string]string {
	ret := make(map[string]string)

	for _, lbl := range lbls {
		for k, v := range lbl {
			ret[k] = v
		}
	}

	return ret
}

func mergeMounts(mounts ...[]corev1.VolumeMount) []corev1.VolumeMount {
	var ret []corev1.VolumeMount

	for _, mnts := range mounts {
		ret = append(ret, mnts...)
	}

	return ret
}

func mergeEnvFromSource(srcs ...[]corev1.EnvFromSource) []corev1.EnvFromSource {
	ret := make([]corev1.EnvFromSource, 0)

	for _, evs := range srcs {
		ret = append(ret, evs...)
	}

	return ret
}

func mergeEnvVar(envs ...[]corev1.EnvVar) []corev1.EnvVar {
	idx := make(map[string]corev1.EnvVar)

	for _, evs := range envs {
		for _, ev := range evs {
			idx[ev.Name] = ev
		}
	}

	keys := make([]string, 0)
	for k := range idx {
		keys = append(keys, k)
	}
	sort.Sort(sort.StringSlice(keys))

	ret := make([]corev1.EnvVar, len(keys))
	for i, k := range keys {
		ret[i] = idx[k]
	}

	return ret
}

func int32Ptr(i int32) *int32 {
	return &i
}
func int64Ptr(i int64) *int64 {
	return &i
}

func boolPtr(t bool) *bool {
	return &t
}

func sandboxNamespace() string {
	return "wasmcloud-operator-system"
}
