package globalrolebinding

import (
	"context"
	"log"
	"reflect"
	"regexp"

	customv1alpha1 "github.com/yagonobre/global-role-binding/pkg/apis/custom/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Add creates a new GLobalRoleBinding Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileGLobalRoleBinding{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

//TODO add namespaces watch
// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("globalrolebinding-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to GLobalRoleBinding
	err = c.Watch(&source.Kind{Type: &customv1alpha1.GLobalRoleBinding{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to RoleBinding
	err = c.Watch(&source.Kind{Type: &rbacv1.RoleBinding{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &customv1alpha1.GLobalRoleBinding{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileGLobalRoleBinding{}

// ReconcileGLobalRoleBinding reconciles a GLobalRoleBinding object
type ReconcileGLobalRoleBinding struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a GLobalRoleBinding object and makes changes based on the state read
// and what is in the GLobalRoleBinding.Spec
// +kubebuilder:rbac:groups=v1,resources=namespaces,verbs=list;watch
// +kubebuilder:rbac:groups=rbac,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=custom.authorization.global.io,resources=globalrolebindings,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileGLobalRoleBinding) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	//When a roleBinding was deleted the request is create with namespace but the resource is unamespaced
	request.NamespacedName.Namespace = ""

	// Fetch the GLobalRoleBinding instance
	instance := &customv1alpha1.GLobalRoleBinding{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	namespaces, err := r.getNamespacesByRegex(instance.Namespaces)
	if err != nil {
		return reconcile.Result{}, err
	}
	for _, namespace := range namespaces {
		if err := r.createOrUpdateRoleBinding(instance, namespace); err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileGLobalRoleBinding) getNamespacesByRegex(namespaceRegex string) ([]string, error) {
	regex := regexp.MustCompile(namespaceRegex)
	namespaceList := &corev1.NamespaceList{}
	result := []string{}
	err := r.List(context.TODO(), nil, namespaceList)
	if err != nil {
		return nil, err
	}
	for _, namespace := range namespaceList.Items {
		if regex.MatchString(namespace.Name) {
			result = append(result, namespace.Name)
		}
	}
	return result, nil
}

func (r *ReconcileGLobalRoleBinding) roleBindingSpec(globalRoleBinding *customv1alpha1.GLobalRoleBinding, namespace string) (*rbacv1.RoleBinding, error) {
	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      globalRoleBinding.Name,
			Namespace: namespace,
		},
		Subjects: []rbacv1.Subject{},
		RoleRef: rbacv1.RoleRef{
			APIGroup: globalRoleBinding.RoleRef.APIGroup,
			Kind:     globalRoleBinding.RoleRef.Kind,
			Name:     globalRoleBinding.RoleRef.Name,
		},
	}

	for _, subject := range globalRoleBinding.Subjects {
		roleBinding.Subjects = append(roleBinding.Subjects, rbacv1.Subject{
			Kind:     subject.Kind,
			APIGroup: subject.APIGroup,
			Name:     subject.Name})
	}

	if err := controllerutil.SetControllerReference(globalRoleBinding, roleBinding, r.scheme); err != nil {
		return nil, err
	}
	return roleBinding, nil
}

//TODO rewrite this function
func (r *ReconcileGLobalRoleBinding) createOrUpdateRoleBinding(globalRoleBinding *customv1alpha1.GLobalRoleBinding, namespace string) error {
	roleBinding, err := r.roleBindingSpec(globalRoleBinding, namespace)
	if err != nil {
		return err
	}
	found := &rbacv1.RoleBinding{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: roleBinding.Name, Namespace: roleBinding.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Printf("Creating RoleBinding %s/%s\n", roleBinding.Namespace, roleBinding.Name)
		err = r.Create(context.TODO(), roleBinding)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	if !reflect.DeepEqual(roleBinding.Subjects, found.Subjects) || !reflect.DeepEqual(roleBinding.RoleRef, found.RoleRef) {
		found = roleBinding
		log.Printf("Updating RoleBinding %s/%s\n", roleBinding.Namespace, roleBinding.Name)
		err = r.Update(context.TODO(), found)
		if err != nil {
			return err
		}
	}
	return nil
}
