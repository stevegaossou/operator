package render_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	operator "github.com/tigera/operator/pkg/apis/operator/v1"
	"github.com/tigera/operator/pkg/render"
	//corev1 "k8s.io/api/core/v1"
	//"k8s.io/apimachinery/pkg/api/resource"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("LogStorage rendering tests", func() {
	var logStorage *operator.LogStorage
	//storageClassName := "tigera-elasticsearch"
	//expectedPvcTemplate := corev1.PersistentVolumeClaim{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name: "elasticsearch-data",
	//	},
	//	Spec: corev1.PersistentVolumeClaimSpec{
	//		AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
	//		Resources: corev1.ResourceRequirements{
	//			Requests: corev1.ResourceList{
	//				"storage": resource.MustParse("10Gi"),
	//			},
	//		},
	//		StorageClassName: &storageClassName,
	//	},
	//}
	BeforeEach(func() {
		// Initialize a default logStorage to use. Each test can override this to its
		// desired configuration.
		logStorage = &operator.LogStorage{
			Spec: operator.LogStorageSpec{
				Certificate: &operator.Certificate{
					SecretName: "tigera-es-config",
				},
				NodeConfig: &operator.NodeConfig{
					NodeCount: 2,
					StorageClass: "",
					ResourceRequirements: nil,
				},
				IndexConfig: &operator.IndexConfig{
					ReplicaCount: 2,
					ShardCount: 2,
				},
			},
			// TODO: is this needed?
			Status: operator.LogStorageStatus{
			},
		}

	})

	It("should render a LogStorageComponent", func() {
		component, err := render.LogStorage(*logStorage)
		resources := component.Objects()
		Expect(len(resources)).To(Equal(2))
		Expect(err).NotTo(HaveOccurred())
		ExpectResource(resources[0], "tigera-elasticsearch", "", "", "", "")
		ExpectResource(resources[1], "tigera-elasticsearch", "tigera-elasticsearch", "elasticsearch.k8s.elastic.co", "v1alpha1", "Elasticsearch")
		//Expect(resources[1].Spec.Nodes[0].VolumeClaimTemplates).To(ConsistOf(expectedPvcTemplate))
	})
})
