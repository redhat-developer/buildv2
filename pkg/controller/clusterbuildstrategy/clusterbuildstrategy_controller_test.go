package clusterbuildstrategy_test

import (
	"context"

	"github.com/k8s-build/build/pkg/config"
	clusterbuildstrategyController "github.com/k8s-build/build/pkg/controller/clusterbuildstrategy"
	"github.com/k8s-build/build/pkg/controller/fakes"
	"github.com/k8s-build/build/pkg/ctxlog"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("Reconcile ClusterBuildStrategy", func() {
	var (
		manager           *fakes.FakeManager
		reconciler        reconcile.Reconciler
		request           reconcile.Request
		buildStrategyName string
	)

	BeforeEach(func() {
		buildStrategyName = "kaniko"

		// Fake the manager and get a reconcile Request
		manager = &fakes.FakeManager{}
		request = reconcile.Request{NamespacedName: types.NamespacedName{Name: buildStrategyName}}
	})

	JustBeforeEach(func() {
		// Reconcile
		testCtx := ctxlog.NewContext(context.TODO(), "fake-logger")
		reconciler = clusterbuildstrategyController.NewReconciler(testCtx, config.NewDefaultConfig(), manager)
	})

	Describe("Reconcile", func() {
		Context("when request a new ClusterBuildStrategy", func() {
			It("succeed without any error", func() {
				result, err := reconciler.Reconcile(request)
				Expect(err).ToNot(HaveOccurred())
				Expect(reconcile.Result{}).To(Equal(result))
			})
		})
	})
})
