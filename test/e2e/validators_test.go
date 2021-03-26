// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubectl/pkg/scheme"

	"github.com/shipwright-io/build/pkg/apis"
	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	"github.com/shipwright-io/build/test/utils"
)

const (
	EnvVarServiceAccountName        = "TEST_E2E_SERVICEACCOUNT_NAME"
	EnvVarVerifyTektonObjects       = "TEST_E2E_VERIFY_TEKTONOBJECTS"
	EnvVarTimeoutMultiplier         = "TEST_E2E_TIMEOUT_MULTIPLIER"
	EnvVarImageRepo                 = "TEST_IMAGE_REPO"
	EnvVarEnablePrivateRepos        = "TEST_PRIVATE_REPO"
	EnvVarImageRepoSecret           = "TEST_IMAGE_REPO_SECRET"
	EnvVarSourceURLGithubSecret     = "TEST_PRIVATE_GITHUB_SECRET"
	EnvVarSourceRepoSecretJSON      = "TEST_IMAGE_REPO_DOCKERCONFIGJSON"
	EnvVarSourceURLGithubSecretJSON = "TEST_PRIVATE_GITHUB_SSHAUTHPRIVATEKEY"
	EnvVarSourceURLGithub           = "TEST_PRIVATE_GITHUB"
	EnvVarSourceURLGitlab           = "TEST_PRIVATE_GITLAB"
	EnvVarSourceURLSecret           = "TEST_SOURCE_SECRET"
)

// createPipelineServiceAccount reads the TEST_E2E_SERVICEACCOUNT_NAME environment variable. If the value is "generated", then nothing is done.
// Otherwise it will create the service account. No error occurs if the service account already exists.
func createPipelineServiceAccount(testBuild *utils.TestBuild) {
	serviceAccountName := os.Getenv(EnvVarServiceAccountName)

	if serviceAccountName == "generated" {
		Logf("Skipping creation of service account, generated one will be used per build run.")
	} else {
		serviceAccount := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: testBuild.Namespace,
				Name:      serviceAccountName,
			},
		}

		Logf("Creating '%s' service-account", serviceAccountName)
		_, err := testBuild.Clientset.CoreV1().ServiceAccounts(testBuild.Namespace).Create(testBuild.Context, serviceAccount, metav1.CreateOptions{})
		if err != nil && !apierrors.IsAlreadyExists(err) {
			Expect(err).ToNot(HaveOccurred(), "Error creating service account")
		}
	}
}

// createContainerRegistrySecret use environment variables to check for container registry
// credentials secret, when not found a new secret is created.
func createContainerRegistrySecret(testBuild *utils.TestBuild) {
	secretName := os.Getenv(EnvVarImageRepoSecret)
	secretPayload := os.Getenv(EnvVarSourceRepoSecretJSON)
	if secretName == "" || secretPayload == "" {
		Logf("Container registry secret won't be created.")
		return
	}

	_, err := testBuild.LookupSecret(types.NamespacedName{Namespace: testBuild.Namespace, Name: secretName})
	if err == nil {
		Logf("Container registry secret is found at '%s/%s'", testBuild.Namespace, secretName)
		return
	}

	payload := []byte(secretPayload)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testBuild.Namespace,
			Name:      secretName,
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			corev1.DockerConfigJsonKey: payload,
		},
	}

	Logf("Creating container-registry secret '%s/%s' (%d bytes)", testBuild.Namespace, secretName, len(payload))
	err = testBuild.CreateSecret(secret)
	Expect(err).ToNot(HaveOccurred(), "on creating container registry secret")
}

// createGithubSSHKeySecret use environment variables to check for container registry
// credentials secret, when not found a new secret is created.
func createGithubSSHKeySecret(testBuild *utils.TestBuild) {
	secretName := os.Getenv(EnvVarSourceURLGithubSecret)
	secretPayload := os.Getenv(EnvVarSourceURLGithubSecretJSON)
	if secretName == "" || secretPayload == "" {
		Logf("secretName: %s, secretPayload: %s", secretName, secretPayload)
		Logf("Private Github repository access secret won't be created.")
		return
	}

	_, err := testBuild.LookupSecret(types.NamespacedName{Namespace: testBuild.Namespace, Name: secretName})
	if err == nil {
		Logf("Private Github access secret is found at '%s/%s'", testBuild.Namespace, secretName)
		return
	}

	payload := []byte(secretPayload)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testBuild.Namespace,
			Name:      secretName,
		},
		Type: corev1.SecretTypeSSHAuth,
		Data: map[string][]byte{
			corev1.SSHAuthPrivateKey: payload,
		},
	}

	Logf("Creating private github access secret '%s/%s' (%d bytes)", testBuild.Namespace, secretName, len(payload))
	err = testBuild.CreateSecret(secret)
	Expect(err).ToNot(HaveOccurred(), "on creating private github access secret")
}

// validateBuildRunToSucceed creates the build run and watches its flow until it succeeds.
func validateBuildRunToSucceed(testBuild *utils.TestBuild, testBuildRun *buildv1alpha1.BuildRun) {
	trueCondition := corev1.ConditionTrue
	falseCondition := corev1.ConditionFalse

	// Ensure the BuildRun has been created
	err := testBuild.CreateBR(testBuildRun)
	Expect(err).ToNot(HaveOccurred(), "Failed to create BuildRun")

	// Ensure a BuildRun eventually moves to a succeeded TRUE status
	nextStatusLog := time.Now().Add(60 * time.Second)
	Eventually(func() corev1.ConditionStatus {
		testBuildRun, err = testBuild.LookupBuildRun(types.NamespacedName{Name: testBuildRun.Name, Namespace: testBuild.Namespace})
		Expect(err).ToNot(HaveOccurred(), "Error retrieving a buildRun")

		if testBuildRun.Status.GetCondition(buildv1alpha1.Succeeded) == nil {
			return corev1.ConditionUnknown
		}

		Expect(testBuildRun.Status.GetCondition(buildv1alpha1.Succeeded).Status).ToNot(Equal(falseCondition), "BuildRun status doesn't move to Succeeded")

		now := time.Now()
		if now.After(nextStatusLog) {
			Logf("Still waiting for build run '%s' to succeed.", testBuildRun.Name)
			nextStatusLog = time.Now().Add(60 * time.Second)
		}

		return testBuildRun.Status.GetCondition(buildv1alpha1.Succeeded).Status

	}, time.Duration(1100*getTimeoutMultiplier())*time.Second, 5*time.Second).Should(Equal(trueCondition), "BuildRun did not succeed")

	// Verify that the BuildSpec is still available in the status
	Expect(testBuildRun.Status.BuildSpec).ToNot(BeNil(), "BuildSpec is not available in the status")

	Logf("Test build '%s' is completed after %v !", testBuildRun.GetName(), testBuildRun.Status.CompletionTime.Time.Sub(testBuildRun.Status.StartTime.Time))
}

// validateServiceAccountDeletion validates that a service account is correctly deleted after the end of
// a build run and depending on the state of the build run
func validateServiceAccountDeletion(buildRun *buildv1alpha1.BuildRun, namespace string) {
	buildRunCondition := buildRun.Status.GetCondition(buildv1alpha1.Succeeded)
	if buildRunCondition != nil {
		if buildRunCondition.Status == "" || buildRunCondition.Status == corev1.ConditionUnknown {
			Logf("Skipping validation of service account deletion because build run did not end.")
			return
		}
	}

	if buildRun.Spec.ServiceAccount == nil || !buildRun.Spec.ServiceAccount.Generate {
		Logf("Skipping validation of service account deletion because service account is not generated")
		return
	}

	saNamespacedName := types.NamespacedName{
		Name:      buildRun.Name + "-sa",
		Namespace: namespace,
	}

	Logf("Verifying that service account '%s' has been deleted.", saNamespacedName.Name)
	_, err := testBuild.LookupServiceAccount(saNamespacedName)
	Expect(err).To(HaveOccurred(), "Expected error to retrieve the generated service account after build run completion.")
	Expect(apierrors.IsNotFound(err)).To(BeTrue(), "Expected service account to be deleted.")
}

func readAndDecode(filePath string) (runtime.Object, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	if err := apis.AddToScheme(scheme.Scheme); err != nil {
		return nil, err
	}

	payload, err := ioutil.ReadFile(filepath.Join("..", "..", filePath))
	if err != nil {
		return nil, err
	}

	obj, _, err := decode(payload, nil, nil)
	return obj, err
}

// buildStrategyTestData gets the us the BuildStrategy test data set up
func buildStrategyTestData(ns string, buildStrategyCRPath string) (*buildv1alpha1.BuildStrategy, error) {
	obj, err := readAndDecode(buildStrategyCRPath)
	if err != nil {
		return nil, err
	}

	buildStrategy := obj.(*buildv1alpha1.BuildStrategy)
	buildStrategy.SetNamespace(ns)

	return buildStrategy, err
}

func buildTestData(namespace string, identifier string, filePath string) (*buildv1alpha1.Build, error) {
	obj, err := readAndDecode(filePath)
	if err != nil {
		return nil, err
	}

	build, ok := obj.(*buildv1alpha1.Build)
	if !ok {
		return nil, fmt.Errorf("failed to use the content of %s as a Build runtime object", filePath)
	}

	build.SetNamespace(namespace)
	build.SetName(identifier)
	return build, nil
}

// buildTestData gets the us the Build test data set up
func buildRunTestData(ns string, identifier string, filePath string) (*buildv1alpha1.BuildRun, error) {
	obj, err := readAndDecode(filePath)
	if err != nil {
		return nil, err
	}

	buildRun, ok := obj.(*buildv1alpha1.BuildRun)
	if !ok {
		return nil, fmt.Errorf("failed to use the content of %s as a BuildRun runtime object", filePath)
	}

	buildRun.SetNamespace(ns)
	buildRun.SetName(identifier)
	buildRun.Spec.BuildRef.Name = identifier

	serviceAccountName := os.Getenv(EnvVarServiceAccountName)
	if serviceAccountName == "generated" {
		buildRun.Spec.ServiceAccount = &buildv1alpha1.ServiceAccount{
			Generate: true,
		}
	} else {
		buildRun.Spec.ServiceAccount = &buildv1alpha1.ServiceAccount{
			Name: &serviceAccountName,
		}
	}

	return buildRun, nil
}

func getTimeoutMultiplier() int64 {
	value := os.Getenv(EnvVarTimeoutMultiplier)
	if value == "" {
		return 1
	}

	intValue, err := strconv.ParseInt(value, 10, 64)
	Expect(err).ToNot(HaveOccurred(), "Failed to parse EnvVarTimeoutMultiplier to integer")
	return intValue
}
