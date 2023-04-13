// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

package sources

import (
	"fmt"
	"strings"

	core "k8s.io/api/core/v1"

	build "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/shipwright-io/build/pkg/config"
	"github.com/shipwright-io/build/pkg/reconciler/buildrun/resources/steps"

	pipeline "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

// AppendBundleStep appends the bundle step to the TaskSpec
func AppendBundleStep(
	cfg *config.Config,
	taskSpec *pipeline.TaskSpec,
	source build.Source,
	buildStrategySteps []build.BuildStep,
	name string,
) {
	// append the result
	taskSpec.Results = append(taskSpec.Results, pipeline.TaskResult{
		Name:        fmt.Sprintf("%s-source-%s-image-digest", prefixParamsResultsVolumes, name),
		Description: "The digest of the bundle image.",
	})

	// initialize the step from the template
	bundleStep := *cfg.BundleContainerTemplate.DeepCopy()

	// add the build-specific details
	bundleStep.Name = fmt.Sprintf("source-%s", name)
	bundleStep.Args = []string{
		"--image", source.BundleContainer.Image,
		"--target", fmt.Sprintf("$(params.%s-%s)", prefixParamsResultsVolumes, paramSourceRoot),
		"--result-file-image-digest", fmt.Sprintf("$(results.%s-source-%s-image-digest.path)", prefixParamsResultsVolumes, name),
	}

	// add credentials mount, if provided
	if source.Credentials != nil {
		AppendSecretVolume(taskSpec, source.Credentials.Name)

		secretMountPath := fmt.Sprintf("/workspace/%s-pull-secret", prefixParamsResultsVolumes)

		// define the volume mount on the container
		bundleStep.VolumeMounts = append(bundleStep.VolumeMounts, core.VolumeMount{
			Name:      SanitizeVolumeNameForSecretName(source.Credentials.Name),
			MountPath: secretMountPath,
			ReadOnly:  true,
		})

		// append the argument
		bundleStep.Args = append(bundleStep.Args,
			"--secret-path", secretMountPath,
		)
	}

	// add prune flag in when prune after pull is configured
	if source.BundleContainer.Prune != nil && *source.BundleContainer.Prune == build.PruneAfterPull {
		bundleStep.Args = append(bundleStep.Args, "--prune")
	}

	steps.UpdateSecurityContext(&bundleStep, buildStrategySteps)

	taskSpec.Steps = append(taskSpec.Steps, bundleStep)
}

// AppendBundleResult append bundle source result to build run
func AppendBundleResult(buildRun *build.BuildRun, name string, results []pipeline.TaskRunResult) {
	imageDigest := findResultValue(results, fmt.Sprintf("%s-source-%s-image-digest", prefixParamsResultsVolumes, name))

	if strings.TrimSpace(imageDigest) != "" {
		buildRun.Status.Sources = append(buildRun.Status.Sources, build.SourceResult{
			Name: name,
			Bundle: &build.BundleSourceResult{
				Digest: imageDigest,
			},
		})
	}
}
