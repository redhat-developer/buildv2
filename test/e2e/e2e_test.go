// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	buildv1alpha1 "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
)

var _ = Describe("For a Kubernetes cluster with Tekton and build installed", func() {
	var (
		testID string
		err    error

		build    *buildv1alpha1.Build
		buildRun *buildv1alpha1.BuildRun
	)

	AfterEach(func() {
		if CurrentGinkgoTestDescription().Failed {
			printTestFailureDebugInfo(testBuild, testBuild.Namespace, testID)

		} else if buildRun != nil {
			validateServiceAccountDeletion(buildRun, testBuild.Namespace)
		}

		if buildRun != nil {
			testBuild.DeleteBR(buildRun.Name)
			buildRun = nil
		}

		if build != nil {
			testBuild.DeleteBuild(build.Name)
			build = nil
		}
	})

	Context("when a Buildah build is defined", func() {

		BeforeEach(func() {
			testID = generateTestID("buildah")

			// create the build definition
			build = createBuild(
				testBuild,
				testID,
				"samples/build/build_buildah_cr.yaml",
			)
		})

		It("successfully runs a build", func() {
			buildRun, err = buildRunTestData(testBuild.Namespace, testID, "samples/buildrun/buildrun_buildah_cr.yaml")
			Expect(err).ToNot(HaveOccurred(), "Error retrieving buildrun test data")

			validateBuildRunToSucceed(testBuild, buildRun)
		})
	})

	Context("when a Buildah build with a contextDir and a custom Dockerfile name is defined", func() {

		BeforeEach(func() {
			testID = generateTestID("buildah-custom-context-dockerfile")

			// create the build definition
			build = createBuild(
				testBuild,
				testID,
				"test/data/build_buildah_cr_custom_context+dockerfile.yaml",
			)
		})

		It("successfully runs a build", func() {
			buildRun, err = buildRunTestData(testBuild.Namespace, testID, "test/data/buildrun_buildah_cr_custom_context+dockerfile.yaml")
			Expect(err).ToNot(HaveOccurred(), "Error retrieving buildrun test data")

			validateBuildRunToSucceed(testBuild, buildRun)
		})
	})

	Context("when a heroku Buildpacks build is defined using a cluster strategy", func() {

		BeforeEach(func() {
			testID = generateTestID("buildpacks-v3-heroku")

			// create the build definition
			build = createBuild(
				testBuild,
				testID,
				"samples/build/build_buildpacks-v3-heroku_cr.yaml",
			)
		})

		It("successfully runs a build", func() {
			buildRun, err = buildRunTestData(testBuild.Namespace, testID, "samples/buildrun/buildrun_buildpacks-v3-heroku_cr.yaml")
			Expect(err).ToNot(HaveOccurred(), "Error retrieving buildrun test data")

			validateBuildRunToSucceed(testBuild, buildRun)
		})
	})

	Context("when a heroku Buildpacks build is defined using a namespaced strategy", func() {
		var buildStrategy *buildv1alpha1.BuildStrategy

		BeforeEach(func() {
			testID = generateTestID("buildpacks-v3-heroku-namespaced")

			buildStrategy, err = buildStrategyTestData(testBuild.Namespace, "samples/buildstrategy/buildpacks-v3/buildstrategy_buildpacks-v3-heroku_namespaced_cr.yaml")
			Expect(err).ToNot(HaveOccurred())

			err = testBuild.CreateBuildStrategy(buildStrategy)
			Expect(err).ToNot(HaveOccurred())

			// create the build definition
			build = createBuild(
				testBuild,
				testID,
				"samples/build/build_buildpacks-v3-heroku_namespaced_cr.yaml",
			)
		})

		It("successfully runs a build", func() {
			buildRun, err = buildRunTestData(testBuild.Namespace, testID, "samples/buildrun/buildrun_buildpacks-v3-heroku_namespaced_cr.yaml")
			Expect(err).ToNot(HaveOccurred())

			validateBuildRunToSucceed(testBuild, buildRun)
		})

		AfterEach(func() {
			err = testBuild.DeleteBuildStrategy(buildStrategy.Name)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("when a Buildpacks v3 build is defined using a cluster strategy", func() {

		BeforeEach(func() {
			testID = generateTestID("buildpacks-v3")

			// create the build definition
			build = createBuild(
				testBuild,
				testID,
				"samples/build/build_buildpacks-v3_cr.yaml",
			)
		})

		It("successfully runs a build", func() {
			buildRun, err = buildRunTestData(testBuild.Namespace, testID, "samples/buildrun/buildrun_buildpacks-v3_cr.yaml")
			Expect(err).ToNot(HaveOccurred(), "Error retrieving buildrun test data")

			validateBuildRunToSucceed(testBuild, buildRun)
		})
	})

	Context("when a Buildpacks v3 build is defined using a namespaced strategy", func() {
		var buildStrategy *buildv1alpha1.BuildStrategy

		BeforeEach(func() {
			testID = generateTestID("buildpacks-v3-namespaced")

			buildStrategy, err = buildStrategyTestData(testBuild.Namespace, "samples/buildstrategy/buildpacks-v3/buildstrategy_buildpacks-v3_namespaced_cr.yaml")
			Expect(err).ToNot(HaveOccurred())

			err = testBuild.CreateBuildStrategy(buildStrategy)
			Expect(err).ToNot(HaveOccurred())

			// create the build definition
			build = createBuild(
				testBuild,
				testID,
				"samples/build/build_buildpacks-v3_namespaced_cr.yaml",
			)
		})

		It("successfully runs a build", func() {
			buildRun, err = buildRunTestData(testBuild.Namespace, testID, "samples/buildrun/buildrun_buildpacks-v3_namespaced_cr.yaml")
			Expect(err).ToNot(HaveOccurred(), "Error retrieving buildrun test data")

			validateBuildRunToSucceed(testBuild, buildRun)
		})

		AfterEach(func() {
			err = testBuild.DeleteBuildStrategy(buildStrategy.Name)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("when a Buildpacks v3 build is defined for a php runtime", func() {

		BeforeEach(func() {
			testID = generateTestID("buildpacks-v3-php")

			// create the build definition
			build = createBuild(
				testBuild,
				testID,
				"test/data/build_buildpacks-v3_php_cr.yaml",
			)
		})

		It("successfully runs a build", func() {
			buildRun, err = buildRunTestData(testBuild.Namespace, testID, "test/data/buildrun_buildpacks-v3_php_cr.yaml")
			Expect(err).ToNot(HaveOccurred(), "Error retrieving buildrun test data")

			validateBuildRunToSucceed(testBuild, buildRun)
		})
	})

	Context("when a Buildpacks v3 build is defined for a ruby runtime", func() {

		BeforeEach(func() {
			testID = generateTestID("buildpacks-v3-ruby")

			// create the build definition
			build = createBuild(
				testBuild,
				testID,
				"test/data/build_buildpacks-v3_ruby_cr.yaml",
			)
		})

		It("successfully runs a build", func() {
			buildRun, err = buildRunTestData(testBuild.Namespace, testID, "test/data/buildrun_buildpacks-v3_ruby_cr.yaml")
			Expect(err).ToNot(HaveOccurred())

			validateBuildRunToSucceed(testBuild, buildRun)
		})
	})

	Context("when a Buildpacks v3 build is defined for a golang runtime", func() {

		BeforeEach(func() {
			testID = generateTestID("buildpacks-v3-golang")

			// create the build definition
			build = createBuild(
				testBuild,
				testID,
				"test/data/build_buildpacks-v3_golang_cr.yaml",
			)
		})

		It("successfully runs a build", func() {
			buildRun, err = buildRunTestData(testBuild.Namespace, testID, "test/data/buildrun_buildpacks-v3_golang_cr.yaml")
			Expect(err).ToNot(HaveOccurred(), "Error retrieving buildrun test data")

			validateBuildRunToSucceed(testBuild, buildRun)
		})
	})

	Context("when a Buildpacks v3 build is defined for a java runtime", func() {

		BeforeEach(func() {
			testID = generateTestID("buildpacks-v3-java")

			// create the build definition
			build = createBuild(
				testBuild,
				testID,
				"test/data/build_buildpacks-v3_java_cr.yaml",
			)
		})

		It("successfully runs a build", func() {
			buildRun, err = buildRunTestData(testBuild.Namespace, testID, "test/data/buildrun_buildpacks-v3_java_cr.yaml")
			Expect(err).ToNot(HaveOccurred())

			validateBuildRunToSucceed(testBuild, buildRun)
		})
	})

	Context("when a buildpacks-v3 build is defined for a nodejs app with runtime-image", func() {

		BeforeEach(func() {
			testID = generateTestID("buildpacks-v3-nodejs-ex-runtime")

			build = createBuild(
				testBuild,
				testID,
				"test/data/build_buildpacks-v3_nodejs_runtime-image_cr.yaml",
			)
		})

		It("successfully runs a build", func() {
			buildRun, err = buildRunTestData(testBuild.Namespace, testID, "test/data/buildrun_buildpacks-v3_nodejs_runtime-image_cr.yaml")
			Expect(err).ToNot(HaveOccurred(), "Error retrieving buildrun test data")

			validateBuildRunToSucceed(testBuild, buildRun)
		})
	})

	Context("when a Kaniko build is defined to use public GitHub", func() {

		BeforeEach(func() {
			testID = generateTestID("kaniko")

			// create the build definition
			build = createBuild(
				testBuild,
				testID,
				"samples/build/build_kaniko_cr.yaml",
			)
		})

		It("successfully runs a build", func() {
			buildRun, err = buildRunTestData(testBuild.Namespace, testID, "samples/buildrun/buildrun_kaniko_cr.yaml")
			Expect(err).ToNot(HaveOccurred())

			validateBuildRunToSucceed(testBuild, buildRun)
		})
	})

	Context("when a Kaniko build with a Dockerfile that requires advanced permissions is defined", func() {

		BeforeEach(func() {
			testID = generateTestID("kaniko-advanced-dockerfile")

			// create the build definition
			build = createBuild(
				testBuild,
				testID,
				"test/data/build_kaniko_cr_advanced_dockerfile.yaml",
			)
		})

		It("successfully runs a build", func() {
			buildRun, err = buildRunTestData(testBuild.Namespace, testID, "test/data/buildrun_kaniko_cr_advanced_dockerfile.yaml")
			Expect(err).ToNot(HaveOccurred(), "Error retrieving buildrun test data")

			validateBuildRunToSucceed(testBuild, buildRun)
		})
	})

	Context("when a Kaniko build with a contextDir and a custom Dockerfile name is defined", func() {

		BeforeEach(func() {
			testID = generateTestID("kaniko-custom-context-dockerfile")

			// create the build definition
			build = createBuild(
				testBuild,
				testID,
				"test/data/build_kaniko_cr_custom_context+dockerfile.yaml",
			)
		})

		It("successfully runs a build", func() {
			buildRun, err = buildRunTestData(testBuild.Namespace, testID, "test/data/buildrun_kaniko_cr_custom_context+dockerfile.yaml")
			Expect(err).ToNot(HaveOccurred())

			validateBuildRunToSucceed(testBuild, buildRun)
		})
	})

	Context("when a Buildkit build with a contextDir and a path to a Dockerfile is defined", func() {

		BeforeEach(func() {
			testID = generateTestID("buildkit-custom-context")

			// create the build definition
			build = createBuild(
				testBuild,
				testID,
				"test/data/build_buildkit_cr_insecure_registry.yaml",
			)
		})

		It("successfully runs a build", func() {
			buildRun, err = buildRunTestData(testBuild.Namespace, testID, "samples/buildrun/buildrun_buildkit_cr.yaml")
			Expect(err).ToNot(HaveOccurred())

			validateBuildRunToSucceed(testBuild, buildRun)
		})
	})

	Context("when a s2i build is defined", func() {

		BeforeEach(func() {
			testID = generateTestID("s2i")

			// create the build definition
			build = createBuild(
				testBuild,
				testID,
				"samples/build/build_source-to-image_cr.yaml",
			)
		})

		It("successfully runs a build", func() {
			buildRun, err = buildRunTestData(testBuild.Namespace, testID, "samples/buildrun/buildrun_source-to-image_cr.yaml")
			Expect(err).ToNot(HaveOccurred(), "Error retrieving buildrun test data")

			validateBuildRunToSucceed(testBuild, buildRun)
		})
	})

	Context("when a private source repository is used", func() {

		BeforeEach(func() {
			if os.Getenv(EnvVarEnablePrivateRepos) != "true" {
				Skip("Skipping test cases that use a private source repository")
			}
		})

		Context("when a nodejs build is defined to use a private GitHub repository", func() {
			BeforeEach(func() {
				testID = generateTestID("private-github-nodejs-buildpack")

				// create the build definition
				build = createBuild(
					testBuild,
					testID,
					"test/data/build_buildpacks-v3_private_github_cr.yaml",
				)
			})

			It("successfully runs a build", func() {
				buildRun, err = buildRunTestData(testBuild.Namespace, testID, "test/data/buildrun_buildpacks-v3_private_github_cr.yaml")
				Expect(err).ToNot(HaveOccurred(), "Error retrieving buildrun test data")

				validateBuildRunToSucceed(testBuild, buildRun)
			})
		})
	})
})
