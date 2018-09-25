package globalrolebinding

import (
	"context"
	"testing"
	"time"

	"github.com/onsi/gomega"
	customv1alpha1 "github.com/yagonobre/global-role-binding/pkg/apis/custom/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var c client.Client

var expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: "test"}}
var expectedRequestWithNamespace = reconcile.Request{NamespacedName: types.NamespacedName{Name: "test", Namespace: "mr-mohr"}}
var rbKey = types.NamespacedName{Name: "test", Namespace: "mr-mohr"}

const timeout = time.Second * 5

var globalRoleBinding = &customv1alpha1.GlobalRoleBinding{
	ObjectMeta: metav1.ObjectMeta{Name: "test"},
	Subjects: []customv1alpha1.Subject{{
		Kind:     "Group",
		APIGroup: "rbac.authorization.k8s.io",
		Name:     "test",
	}},
	RoleRef: customv1alpha1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     "test",
	},
	Namespaces: "mr-*",
}

var clusterRole = &rbacv1.ClusterRole{
	ObjectMeta: metav1.ObjectMeta{Name: "test"},
	Rules: []rbacv1.PolicyRule{{
		Verbs:     []string{"get"},
		APIGroups: []string{"apps"},
		Resources: []string{"deployments"},
	}},
}

func TestReconcile(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c = mgr.GetClient()

	ns1 := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "mr-mohr"}}
	err = c.Create(context.TODO(), ns1)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), ns1)

	ns2 := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "baby-mohr"}}
	err = c.Create(context.TODO(), ns2)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), ns2)

	err = c.Create(context.TODO(), clusterRole)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), clusterRole)

	recFn, requests := SetupTestReconcile(newReconciler(mgr))
	g.Expect(add(mgr, recFn)).NotTo(gomega.HaveOccurred())
	defer close(StartTestManager(mgr, g))

	// Create the GlobalRoleBinding object and expect the Reconcile and RoleBinding to be created
	err = c.Create(context.TODO(), globalRoleBinding)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), globalRoleBinding)
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))

	roleBinding := &rbacv1.RoleBinding{}
	g.Eventually(func() error { return c.Get(context.TODO(), rbKey, roleBinding) }, timeout).
		Should(gomega.Succeed())

	// Delete the RoleBinding and expect Reconcile to be called for RoleBinding deletion
	g.Expect(c.Delete(context.TODO(), roleBinding)).NotTo(gomega.HaveOccurred())
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequestWithNamespace)))
	g.Eventually(func() error { return c.Get(context.TODO(), rbKey, roleBinding) }, timeout).
		Should(gomega.Succeed())

	// Manually delete RoleBinding since GC isn't enabled in the test control plane
	g.Expect(c.Delete(context.TODO(), roleBinding)).To(gomega.Succeed())
}
