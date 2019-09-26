package render_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	operator "github.com/tigera/operator/pkg/apis/operator/v1"
	"github.com/tigera/operator/pkg/render"
)

var _ = Describe("Logstorage rendering tests", func() {
	var logStorage *operator.LogStorage
	BeforeEach(func() {
		// Initialize a default logStorage to use. Each test can override this to its
		// desired configuration.
		logStorage = &operator.LogStorage{
			Spec: &operator.LogStorageSpec{
				Certificate: &operator.Certificate{
					// TODO: determine correct secret name
					SecretName: "tigera-es-config",
				},
				NodeConfig: &operator.NodeConfig{
					NodeCount: 2,
					StorageClass: "",
					// TODO: determine which value to use
					ResourceRequirements: nil,
				},
				IndexConfig: &operator.IndexConfig{
					ReplicaCount: 2,
					ShardCount: 2,
				},
			},
			// TODO: determine if needed
			Status: &operator.LogStorageStatus{
			},
		}

	})

	It("should render a LogStorageComponent", func() {
		component := render.LogStorage(logStorage)
		resources := component.Objects()
		Expect(len(resources)).To(Equal(2))
		ExpectResource(resources[0], "tigera-elasticsearch", "", "", "v1", "StorageClass")
		ExpectResource(resources[1], "tigera-elasticsearch", "tigera-elasticsearch", "", "v1", "Elasticsearch")
	})
})
