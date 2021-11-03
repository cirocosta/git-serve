package v1alpha1

import (
	"github.com/vmware-labs/reconciler-runtime/apis"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	GitServerConditionReady                              = apis.ConditionReady
	GitServerConditionDeploymentReady apis.ConditionType = "DeploymentReady"
	GitServerConditionServiceReady    apis.ConditionType = "ServiceReady"
	GitServerConditionSecretReady     apis.ConditionType = "SecretReady"
)

var gitServerCondSet = apis.NewLivingConditionSet(
	GitServerConditionDeploymentReady,
	GitServerConditionServiceReady,
)

func (s *GitServerStatus) GetObservedGeneration() int64 {
	return s.ObservedGeneration
}

func (s *GitServerStatus) IsReady() bool {
	return gitServerCondSet.Manage(s).IsHappy()
}

func (*GitServerStatus) GetReadyConditionType() apis.ConditionType {
	return GitServerConditionReady
}

func (s *GitServerStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return gitServerCondSet.Manage(s).GetCondition(t)
}

func (s *GitServerStatus) InitializeConditions() {
	gitServerCondSet.Manage(s).InitializeConditions()
}

func (s *GitServerStatus) PropagateDeploymentStatus(cds *appsv1.DeploymentStatus) {
	var available, progressing *appsv1.DeploymentCondition
	for i := range cds.Conditions {
		switch cds.Conditions[i].Type {
		case appsv1.DeploymentAvailable:
			available = &cds.Conditions[i]
		case appsv1.DeploymentProgressing:
			progressing = &cds.Conditions[i]
		}
	}
	if available == nil || progressing == nil {
		return
	}
	if progressing.Status == corev1.ConditionTrue && available.Status == corev1.ConditionFalse {
		// DeploymentAvailable is False while progressing, avoid reporting GitServerConditionReady as False
		gitServerCondSet.Manage(s).MarkUnknown(GitServerConditionDeploymentReady, progressing.Reason, progressing.Message)
		return
	}
	switch {
	case available.Status == corev1.ConditionUnknown:
		gitServerCondSet.Manage(s).MarkUnknown(GitServerConditionDeploymentReady, available.Reason, available.Message)
	case available.Status == corev1.ConditionTrue:
		gitServerCondSet.Manage(s).MarkTrue(GitServerConditionDeploymentReady)
	case available.Status == corev1.ConditionFalse:
		gitServerCondSet.Manage(s).MarkFalse(GitServerConditionDeploymentReady, available.Reason, available.Message)
	}
}

func (s *GitServerStatus) PropagateServiceStatus(ss *corev1.ServiceStatus) {
	// services don't have meaningful status
	gitServerCondSet.Manage(s).MarkTrue(GitServerConditionServiceReady)
}

func (s *GitServerStatus) PropagateSecretStatus() {
	// secrets don't have meaningful status
	gitServerCondSet.Manage(s).MarkTrue(GitServerConditionSecretReady)
}
