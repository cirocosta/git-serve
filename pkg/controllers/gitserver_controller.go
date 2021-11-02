package controllers

import (
	"context"
	"fmt"

	"github.com/vmware-labs/reconciler-runtime/reconcilers"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"

	v1alpha1 "github.com/cirocosta/git-serve/pkg/apis/v1alpha1"
)

func GitServerReconciler(c reconcilers.Config) *reconcilers.ParentReconciler {
	return &reconcilers.ParentReconciler{
		Type: &v1alpha1.GitServer{},
		Reconciler: reconcilers.Sequence{
			GitServerChildServiceReconciler(c),
			GitServerChildDeploymentReconciler(c),
		},

		Config: c,
	}
}

func GitServerChildDeploymentReconciler(c reconcilers.Config) reconcilers.SubReconciler {
	c.Log = c.Log.WithName("childdeployment")

	return &reconcilers.ChildReconciler{
		Config: c,

		ChildType:     &appsv1.Deployment{},
		ChildListType: &appsv1.DeploymentList{},

		DesiredChild: GitServerDesiredDeploymentChild,

		ReflectChildStatusOnParent: func(parent *v1alpha1.GitServer, child *appsv1.Deployment, err error) {
			if child == nil {
				parent.Status.DeploymentRef = nil
				return
			}

			parent.Status.DeploymentRef = v1alpha1.
				NewTypedLocalObjectReferenceForObject(
					child, c.Scheme(),
				)

			parent.Status.PropagateDeploymentStatus(&child.Status)
		},

		HarmonizeImmutableFields: func(current, desired *appsv1.Deployment) {
			desired.Spec.Replicas = current.Spec.Replicas
		},

		MergeBeforeUpdate: func(current, desired *appsv1.Deployment) {
			current.Labels = desired.Labels
			current.Spec = desired.Spec
		},

		SemanticEquals: func(a1, a2 *appsv1.Deployment) bool {
			return equality.Semantic.DeepEqual(a1.Spec, a2.Spec) &&
				equality.Semantic.DeepEqual(a1.Labels, a2.Labels)
		},

		Sanitize: func(child *appsv1.Deployment) interface{} {
			return child.Spec
		},
	}
}

func GitServerDesiredDeploymentChild(
	ctx context.Context, parent *v1alpha1.GitServer,
) (*appsv1.Deployment, error) {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: appsv1.SchemeGroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: make(map[string]string),
			Name:        parent.Name,
			Namespace:   parent.Namespace,
			Labels:      GitServerLabel(parent.Name),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: GitServerLabel(parent.Name),
			},
			RevisionHistoryLimit: pointer.Int32Ptr(0),
			Template:             GitServerPodTemplateSpec(parent),
		},
	}, nil
}

func GitServerChildServiceReconciler(c reconcilers.Config) reconcilers.SubReconciler {
	c.Log = c.Log.WithName("servicechild")

	return &reconcilers.ChildReconciler{
		Config: c,

		ChildType:     &corev1.Service{},
		ChildListType: &corev1.ServiceList{},

		DesiredChild: GitServerServiceDesiredChild,

		ReflectChildStatusOnParent: func(parent *v1alpha1.GitServer, child *corev1.Service, err error) {
			if child == nil {
				parent.Status.ServiceRef = nil
				parent.Status.Address = nil
				return
			}

			parent.Status.ServiceRef = v1alpha1.
				NewTypedLocalObjectReferenceForObject(
					child, c.Scheme(),
				)

			url := fmt.Sprintf("http://%s.%s.%s",
				child.Name, child.Namespace,
				"svc.cluster.local",
			)

			parent.Status.Address = &v1alpha1.Addressable{URL: url}
			parent.Status.PropagateServiceStatus(&child.Status)
		},

		HarmonizeImmutableFields: func(current, desired *corev1.Service) {
			desired.Spec.ClusterIP = current.Spec.ClusterIP
		},

		MergeBeforeUpdate: func(current, desired *corev1.Service) {
			current.Labels = desired.Labels
			current.Spec = desired.Spec
		},

		SemanticEquals: func(a1, a2 *corev1.Service) bool {
			return equality.Semantic.DeepEqual(a1.Spec, a2.Spec) &&
				equality.Semantic.DeepEqual(a1.Labels, a2.Labels)
		},

		Sanitize: func(child *corev1.Service) interface{} {
			return child.Spec
		},
	}
}

func GitServerServiceDesiredChild(
	ctx context.Context, parent *v1alpha1.GitServer,
) (*corev1.Service, error) {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: corev1.SchemeGroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: make(map[string]string),
			Name:        parent.Name,
			Namespace:   parent.Namespace,
			Labels:      GitServerLabel(parent.Name),
		},
		Spec: corev1.ServiceSpec{
			Selector: GitServerLabel(parent.ObjectMeta.Name),
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       int32(80),
					TargetPort: intstr.FromInt(int(8080)),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "ssh",
					Port:       int32(22),
					TargetPort: intstr.FromInt(int(2222)),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}, nil
}

func GitServerLabel(name string) map[string]string {
	return map[string]string{
		v1alpha1.GroupVersion.Group + "/git-server": name,
	}
}

func GitServerPodTemplateSpec(parent *v1alpha1.GitServer) corev1.PodTemplateSpec {
	args := []string{
		"git-serve",
		"-v",
		"-data-dir=/git-repositories",
		"-ssh-no-auth",
		"-http-no-auth",
	}

	container := corev1.Container{
		Name:  "git-serve",
		Image: parent.Spec.Image,
		Args:  args,
	}

	template := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: GitServerLabel(parent.Name),
		},
		Spec: corev1.PodSpec{
			TerminationGracePeriodSeconds: pointer.Int64Ptr(60),
			Containers: []corev1.Container{
				container,
			},
		},
	}

	return template
}
