package controllers

import (
	"context"
	"fmt"

	"github.com/vmware-labs/reconciler-runtime/reconcilers"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"

	"github.com/cirocosta/git-serve/pkg"
	v1alpha1 "github.com/cirocosta/git-serve/pkg/apis/v1alpha1"
)

const (
	GitServerSSHDataKeyPrivateKey     = "ssh-privatekey"
	GitServerSSHDataKeyPublicKey      = "ssh-publickey"
	GitServerSSHDataKeyAuthorizedKeys = "ssh-authorizedkeys"
	GitServerSSHDataKeyKnownHosts     = "known_hosts"
)

func GitServerReconciler(c reconcilers.Config, defaultImage string) *reconcilers.ParentReconciler {
	return &reconcilers.ParentReconciler{
		Type: &v1alpha1.GitServer{},
		Reconciler: reconcilers.Sequence{
			GitServerDefaultsReconciler(c, defaultImage),
			GitServerChildSecretSyncReconciler(c),
			GitServerChildSecretSyncParentSpecReconciler(c),
			GitServerChildSecretReconciler(c),
			GitServerChildServiceReconciler(c),
			GitServerChildDeploymentReconciler(c),
		},

		Config: c,
	}
}

func GitServerDefaultsReconciler(c reconcilers.Config, defaultImage string) reconcilers.SubReconciler {
	c.Log = c.Log.WithName("defaults-reconciler")

	return &reconcilers.SyncReconciler{
		Sync: func(ctx context.Context, parent *v1alpha1.GitServer) error {
			if parent.Spec.Image != "" {
				return nil
			}

			parent.Spec.Image = defaultImage

			return nil
		},
	}
}

func SecretFieldValueFetcher(c reconcilers.Config) func(context.Context, v1alpha1.SecretKeyRef, string) ([]byte, error) {
	return func(ctx context.Context, keyRef v1alpha1.SecretKeyRef, ns string) ([]byte, error) {
		secret := &corev1.Secret{}

		if err := c.Get(ctx, types.NamespacedName{
			Name:      keyRef.Name,
			Namespace: ns,
		}, secret); err != nil {
			return nil, fmt.Errorf("get secret '%s': %w",
				keyRef.Name, err,
			)
		}

		if secret.Data == nil {
			return nil, nil
		}

		return secret.Data[keyRef.Key], nil
	}
}

func GitServerChildSecretSyncReconciler(c reconcilers.Config) reconcilers.SubReconciler {
	return &reconcilers.SyncReconciler{
		Sync: func(ctx context.Context, parent *v1alpha1.GitServer) error {
			if parent.Status.SecretRef == nil {
				return nil
			}

			secret := &corev1.Secret{}
			if err := c.Get(ctx, types.NamespacedName{
				Name:      parent.Status.SecretRef.Name,
				Namespace: parent.Namespace,
			}, secret); err != nil {
				if !kerrors.IsNotFound(err) {
					return fmt.Errorf("get secret '%s': %w",
						parent.Status.SecretRef.Name, err,
					)
				}

				return nil
			}

			StashSecretData(ctx, secret.Data)
			return nil
		},
	}
}

func GitServerChildSecretSyncParentSpecReconciler(c reconcilers.Config) reconcilers.SubReconciler {
	keyRefFetcher := SecretFieldValueFetcher(c)

	return &reconcilers.SyncReconciler{
		Sync: func(ctx context.Context, parent *v1alpha1.GitServer) error {
			data := RetrieveSecretData(ctx)

			if parent.Spec.SSH != nil {
				if parent.Spec.SSH.Auth.AuthorizedKeys != nil {
					key, err := keyRefFetcher(ctx, parent.Spec.SSH.Auth.AuthorizedKeys.ValueFrom.SecretKeyRef, parent.Namespace)
					if err != nil {
						return err
					}

					data[GitServerSSHDataKeyAuthorizedKeys] = key
				}

				if parent.Spec.SSH.Auth.HostKey != nil {
					key, err := keyRefFetcher(ctx, parent.Spec.SSH.Auth.HostKey.ValueFrom.SecretKeyRef, parent.Namespace)
					if err != nil {
						return err
					}

					data[GitServerSSHDataKeyPrivateKey] = key
				}
			}

			StashSecretData(ctx, data)

			return nil
		},
	}
}

func GitServerChildSecretReconciler(c reconcilers.Config) reconcilers.SubReconciler {
	c.Log = c.Log.WithName("child-secret")

	return &reconcilers.ChildReconciler{
		Config: c,

		ChildType:     &corev1.Secret{},
		ChildListType: &corev1.SecretList{},

		DesiredChild: GitServerDesiredSecretChild,

		ReflectChildStatusOnParent: func(parent *v1alpha1.GitServer, child *corev1.Secret, err error) {
			if child == nil {
				parent.Status.SecretRef = nil
				return
			}

			parent.Status.SecretRef = v1alpha1.
				NewTypedLocalObjectReferenceForObject(
					child, c.Scheme(),
				)

			parent.Status.PropagateSecretStatus()
		},

		HarmonizeImmutableFields: func(current, desired *corev1.Secret) {
		},

		MergeBeforeUpdate: func(current, desired *corev1.Secret) {
			current.Labels = desired.Labels
			current.Data = desired.Data
		},

		SemanticEquals: func(a1, a2 *corev1.Secret) bool {
			return equality.Semantic.DeepEqual(a1.Data, a2.Data) &&
				equality.Semantic.DeepEqual(a1.Labels, a2.Labels)
		},

		Sanitize: func(child *corev1.Secret) interface{} {
			return child.Data
		},
	}
}

func GitServerDesiredSecretChild(
	ctx context.Context, parent *v1alpha1.GitServer,
) (*corev1.Secret, error) {
	data := RetrieveSecretData(ctx)

	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: corev1.SchemeGroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: make(map[string]string),
			Name:        parent.Name,
			Namespace:   parent.Namespace,
			Labels:      GitServerLabel(parent.Name),
		},
		Type: corev1.SecretTypeSSHAuth,
		Data: data,
	}

	var generatedPriv, generatedPub []byte
	var err error

	if v, found := secret.Data[GitServerSSHDataKeyPrivateKey]; !found || len(v) == 0 {
		generatedPriv, generatedPub, err = pkg.GenSSHKeyPair()
		if err != nil {
			return nil, fmt.Errorf("gen ssh key pair: %w", err)
		}

		secret.Data[GitServerSSHDataKeyPrivateKey] = generatedPriv
		secret.Data[GitServerSSHDataKeyPublicKey] = generatedPub
		secret.Data[GitServerSSHDataKeyKnownHosts] = []byte(fmt.Sprintf(
			"%s %s%s %s",
			parent.Name, string(generatedPub),
			GitServerAddress(parent), string(generatedPub)),
		)
	}

	if v, found := secret.Data[GitServerSSHDataKeyKnownHosts]; !found || len(v) == 0 {
		pub, err := pkg.DerivePublicFromPrivate(secret.Data[GitServerSSHDataKeyPrivateKey])
		if err != nil {
			return nil, fmt.Errorf("derive pub from priv: %w", err)
		}

		secret.Data[GitServerSSHDataKeyPublicKey] = pub

		secret.Data[GitServerSSHDataKeyKnownHosts] = []byte(fmt.Sprintf(
			"%s %s%s %s",
			parent.Name, string(pub),
			GitServerAddress(parent), string(pub)),
		)
	}

	if v, found := secret.Data[GitServerSSHDataKeyAuthorizedKeys]; !found || len(v) == 0 {
		// by default, authorize itself
		secret.Data[GitServerSSHDataKeyAuthorizedKeys] = secret.Data[GitServerSSHDataKeyPublicKey]
	}

	return secret, nil
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

func GitServerAddress(parent *v1alpha1.GitServer) string {
	return fmt.Sprintf("%s.%s.%s",
		parent.Name, parent.Namespace,
		"svc.cluster.local",
	)
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

			parent.Status.Address = &v1alpha1.Addressable{
				URL: "http://" + GitServerAddress(parent),
			}

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
	sshSecretVolume := corev1.Volume{
		Name: "ssh",
		VolumeSource: corev1.VolumeSource{
			Projected: &corev1.ProjectedVolumeSource{
				Sources: []corev1.VolumeProjection{
					{
						Secret: &corev1.SecretProjection{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: parent.Name,
							},
						},
					},
				},
			},
		},
	}

	gitDataVolume := corev1.Volume{
		Name: "git-data",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}

	container := corev1.Container{
		Name:            "git-serve",
		Image:           parent.Spec.Image,
		ImagePullPolicy: corev1.PullAlways,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      sshSecretVolume.Name,
				MountPath: "/ssh-secret",
			},
			{
				Name:      gitDataVolume.Name,
				MountPath: "/git-data",
			},
		},
	}

	container.Args = []string{
		"git-serve",
		"-v",
		"-data-dir=/git-data",
		"-ssh-host-key=/ssh-secret/ssh-privatekey",
	}

	if parent.Spec.SSH == nil {
		container.Args = append(container.Args, "-ssh-no-auth")
	}

	if parent.Spec.HTTP == nil {
		container.Args = append(container.Args, "-http-no-auth")
	}

	if parent.Spec.SSH != nil {
		sshauth := parent.Spec.SSH.Auth
		if name := sshauth.AuthorizedKeys.ValueFrom.SecretKeyRef.Name; name != "" {
			container.Args = append(container.Args,
				"-ssh-authorized-keys=/ssh-secret/"+GitServerSSHDataKeyAuthorizedKeys,
			)
		}
	}

	if parent.Spec.HTTP != nil {
		usernameRef := parent.Spec.HTTP.Auth.Username.ValueFrom.SecretKeyRef
		passwordRef := parent.Spec.HTTP.Auth.Password.ValueFrom.SecretKeyRef

		container.Env = append(container.Env, []corev1.EnvVar{
			{
				Name: "GIT_SERVE_HTTP_USERNAME",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: usernameRef.Name,
						},
						Key: usernameRef.Key,
					},
				},
			},
			{
				Name: "GIT_SERVE_HTTP_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: passwordRef.Name,
						},
						Key: passwordRef.Key,
					},
				},
			},
		}...)
	}

	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: GitServerLabel(parent.Name),
		},
		Spec: corev1.PodSpec{
			TerminationGracePeriodSeconds: pointer.Int64Ptr(60),
			Volumes: []corev1.Volume{
				sshSecretVolume,
				gitDataVolume,
			},
			Containers: []corev1.Container{
				container,
			},
		},
	}
}

const SecretDataStashKey reconcilers.StashKey = v1alpha1.Group + "/secret-data"

func StashSecretData(ctx context.Context, data map[string][]byte) {
	reconcilers.StashValue(ctx, SecretDataStashKey, data)
}

func RetrieveSecretData(ctx context.Context) map[string][]byte {
	data, ok := reconcilers.RetrieveValue(ctx, SecretDataStashKey).(map[string][]byte)
	if !ok {
		return map[string][]byte{}
	}

	return data
}
