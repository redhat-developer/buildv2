// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

package sources

import (
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"
)

const (
	prefixParamsResultsVolumes = "shp-"

	paramSourceRoot = "source-root"
)

var (
	// secrets are volumes and volumes are mounted as root, as we run as non-root, we must use 0444 to allow non-root to read it
	secretMountMode = pointer.Int32Ptr(0444)
)

// AppendSecretVolume checks if a volume for a secret already exists, if not it appends it to the TaskSpec
func AppendSecretVolume(
	taskSpec *tektonv1beta1.TaskSpec,
	secretName string,
) {
	volumeName := prefixParamsResultsVolumes + secretName

	// ensure we do not add the secret twice
	for _, volume := range taskSpec.Volumes {
		if volume.VolumeSource.Secret != nil && volume.Name == volumeName {
			return
		}
	}

	// append volume for secret
	taskSpec.Volumes = append(taskSpec.Volumes, corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName:  secretName,
				DefaultMode: secretMountMode,
			},
		},
	})
}
