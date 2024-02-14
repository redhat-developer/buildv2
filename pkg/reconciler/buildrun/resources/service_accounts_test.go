// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

package resources_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	buildv1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"
	"github.com/shipwright-io/build/pkg/controller/fakes"
	"github.com/shipwright-io/build/pkg/reconciler/buildrun/resources"
	test "github.com/shipwright-io/build/test/v1beta1_samples"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	crc "sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Operating service accounts", func() {
	var (
		client                  *fakes.FakeClient
		ctl                     test.Catalog
		buildName, buildRunName string
		buildRunSample          *buildv1beta1.BuildRun
	)

	BeforeEach(func() {
		// init vars
		buildName = "foobuild"
		buildRunName = "foobuildrun"
		client = &fakes.FakeClient{}
		buildRunSample = ctl.DefaultBuildRun(buildRunName, buildName)
	})

	// stub client GET calls and return a initialized sa when asking for a sa
	var generateGetSAStub = func(saName string) func(context context.Context, nn types.NamespacedName, object crc.Object, getOptions ...crc.GetOption) error {
		return func(_ context.Context, nn types.NamespacedName, object crc.Object, getOptions ...crc.GetOption) error {
			switch object := object.(type) {
			case *corev1.ServiceAccount:
				ctl.DefaultServiceAccount(saName).DeepCopyInto(object)
				return nil
			}
			return k8serrors.NewNotFound(schema.GroupResource{}, nn.Name)
		}
	}

	// stub client GET calls and return an error when asking for a service account
	var generateGetSAStubWithError = func(customError error) func(context context.Context, nn types.NamespacedName, object crc.Object, getOptions ...crc.GetOption) error {
		return func(_ context.Context, nn types.NamespacedName, object crc.Object, getOptions ...crc.GetOption) error {
			switch object.(type) {
			case *corev1.ServiceAccount:
				return customError
			}
			return k8serrors.NewNotFound(schema.GroupResource{}, nn.Name)
		}
	}

	Context("Retrieving specified service accounts", func() {

		It("should return a modified sa with a secret reference", func() {
			buildRunSample := ctl.BuildRunWithSA(buildRunName, buildName, "foobarsa")

			// stub a GET API call for a service account
			client.GetCalls(generateGetSAStub("foobarsa"))

			sa, err := resources.RetrieveServiceAccount(context.TODO(), client, ctl.BuildWithOutputSecret(buildName, "default", "foosecret"), buildRunSample)
			Expect(err).To(BeNil())
			Expect(len(sa.Secrets)).To(Equal(1))
			Expect(sa.Secrets[0].Name).To(Equal("foosecret"))
		})
		It("should return a namespace default sa with a secret reference", func() {
			buildRunSample := ctl.BuildRunWithoutSA(buildRunName, buildName)

			// stub a GET API call for a service account
			client.GetCalls(generateGetSAStub("default"))

			sa, err := resources.RetrieveServiceAccount(context.TODO(), client, ctl.BuildWithOutputSecret(buildName, "default", "foosecret"), buildRunSample)
			Expect(err).To(BeNil())
			Expect(len(sa.Secrets)).To(Equal(1))
			Expect(sa.Secrets[0].Name).To(Equal("foosecret"))
		})

		It("should return an error if the specified sa is not found", func() {
			buildRunSample := ctl.BuildRunWithSA(buildRunName, buildName, "foobarsa")

			client.GetCalls(generateGetSAStubWithError(k8serrors.NewNotFound(schema.GroupResource{}, "")))

			client.StatusCalls(func() crc.StatusWriter {
				statusWriter := &fakes.FakeStatusWriter{}
				statusWriter.UpdateCalls(func(ctx context.Context, object crc.Object, _ ...crc.SubResourceUpdateOption) error {
					return nil
				})
				return statusWriter
			})

			sa, err := resources.RetrieveServiceAccount(context.TODO(), client, ctl.BuildWithOutputSecret(buildName, "default", "foosecret"), buildRunSample)
			Expect(sa).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(k8serrors.IsNotFound(err)).To(BeTrue())
		})

		It("should return multiple errors if the specified sa is not found and the condition update is not working", func() {
			buildRunSample := ctl.BuildRunWithSA(buildRunName, buildName, "foobarsa")

			client.GetCalls(generateGetSAStubWithError(k8serrors.NewNotFound(schema.GroupResource{}, "")))

			client.StatusCalls(func() crc.StatusWriter {
				statusWriter := &fakes.FakeStatusWriter{}
				statusWriter.UpdateCalls(func(_ context.Context, object crc.Object, _ ...crc.SubResourceUpdateOption) error {
					switch object.(type) {
					case *buildv1beta1.BuildRun:
						return fmt.Errorf("failed")
					}
					return nil
				})
				return statusWriter
			})

			sa, err := resources.RetrieveServiceAccount(context.TODO(), client, ctl.BuildWithOutputSecret(buildName, "default", "foosecret"), buildRunSample)
			Expect(sa).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(resources.IsClientStatusUpdateError(err)).To(BeTrue())
		})
	})

	Context("Retrieving autogenerated service accounts when the spec.serviceAccount .generate value is used", func() {

		It("should provide a generated sa name", func() {
			Expect(resources.GetGeneratedServiceAccountName(buildRunSample)).To(Equal(buildRunSample.Name))
		})

		It("should return a generated sa with a label, ownerreference and a ref secret if it does not exists", func() {
			buildRunSample := ctl.BuildRunWithSAGenerate(buildRunName, buildName)

			// stub a GET API call for a service account
			client.GetCalls(generateGetSAStubWithError(k8serrors.NewNotFound(schema.GroupResource{}, "foobar")))

			mountTokenVal := false

			client.CreateCalls(func(_ context.Context, object crc.Object, _ ...crc.CreateOption) error {
				switch object := object.(type) {
				case *corev1.ServiceAccount:
					Expect(len(object.Secrets)).To(Equal(1))
					Expect(len(object.OwnerReferences)).To(Equal(1))
					Expect(object.Labels[buildv1beta1.LabelBuildRun]).To(Equal(buildRunName))
					Expect(object.Secrets[0].Name).To(Equal("foosecret"))
					Expect(object.AutomountServiceAccountToken).To(Equal(&mountTokenVal))
				}
				return nil
			})

			_, err := resources.RetrieveServiceAccount(context.TODO(), client, ctl.BuildWithOutputSecret(buildName, "default", "foosecret"), buildRunSample)
			Expect(err).To(BeNil())

		})
		It("should return an existing sa and not generate it again if already exists", func() {
			buildRunSample := ctl.BuildRunWithSAGenerate(buildRunName, buildName)

			// stub a GET API call for a service account
			client.GetCalls(generateGetSAStub(buildRunName))

			_, err := resources.RetrieveServiceAccount(context.TODO(), client, ctl.BuildWithOutputSecret(buildName, "default", "foosecret"), buildRunSample)
			Expect(err).To(BeNil())
		})
		It("should return an error if the sa automatic generation fails", func() {
			buildRunSample := ctl.BuildRunWithSAGenerate(buildRunName, buildName)

			// stub a GET API call for a service account
			// fake the calls with the above stub
			client.GetCalls(generateGetSAStubWithError(fmt.Errorf("something wrong happened")))

			_, err := resources.RetrieveServiceAccount(context.TODO(), client, ctl.BuildWithOutputSecret(buildName, "default", "foosecret"), buildRunSample)
			Expect(err).ToNot(BeNil())
		})
	})
})
