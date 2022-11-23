package keptnworkloadinstance

import (
	"context"

	klcv1alpha1 "github.com/keptn/lifecycle-toolkit/operator/api/v1alpha1"
	"github.com/keptn/lifecycle-toolkit/operator/api/v1alpha1/common"
	controllercommon "github.com/keptn/lifecycle-toolkit/operator/controllers/common"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *KeptnWorkloadInstanceReconciler) reconcileDeployment(ctx context.Context, workloadInstance *klcv1alpha1.KeptnWorkloadInstance) (common.KeptnState, error) {
	var isRunning bool
	var err error

	switch workloadInstance.Spec.ResourceReference.Kind {
	case "Pod":
		isRunning, err = r.isPodRunning(ctx, workloadInstance.Spec.ResourceReference, workloadInstance.Namespace)
	case "ReplicaSet":
		isRunning, err = r.isReplicaSetRunning(ctx, workloadInstance.Spec.ResourceReference, workloadInstance.Namespace)
	case "StatefulSet":
		isRunning, err = r.isStatefulSetRunning(ctx, workloadInstance.Spec.ResourceReference, workloadInstance.Namespace)
	case "DaemonSet":
		isRunning, err = r.isDaemonSetRunning(ctx, workloadInstance.Spec.ResourceReference, workloadInstance.Namespace)
	default:
		isRunning, err = false, controllercommon.ErrUnsupportedWorkloadInstanceResourceReference
	}

	if err != nil {
		return common.StateUnknown, err
	}
	if isRunning {
		workloadInstance.Status.DeploymentStatus = common.StateSucceeded
	} else {
		workloadInstance.Status.DeploymentStatus = common.StateProgressing
	}

	err = r.Client.Status().Update(ctx, workloadInstance)
	if err != nil {
		return common.StateUnknown, err
	}
	return workloadInstance.Status.DeploymentStatus, nil
}

func (r *KeptnWorkloadInstanceReconciler) isReplicaSetRunning(ctx context.Context, resource klcv1alpha1.ResourceReference, namespace string) (bool, error) {
	rep := appsv1.ReplicaSet{}
	err := r.Client.Get(ctx, types.NamespacedName{Name: resource.Name, Namespace: namespace}, &rep)
	if err != nil {
		return false, err
	}
	return *rep.Spec.Replicas == rep.Status.AvailableReplicas, nil
}

func (r *KeptnWorkloadInstanceReconciler) isDaemonSetRunning(ctx context.Context, resource klcv1alpha1.ResourceReference, namespace string) (bool, error) {
	daemonSet := &appsv1.DaemonSet{}
	err := r.Client.Get(ctx, types.NamespacedName{Name: resource.Name, Namespace: namespace}, daemonSet)
	if err != nil {
		return false, err
	}
	return daemonSet.Status.DesiredNumberScheduled == daemonSet.Status.NumberReady, nil
}

func (r *KeptnWorkloadInstanceReconciler) isPodRunning(ctx context.Context, resource klcv1alpha1.ResourceReference, namespace string) (bool, error) {
	podList := &corev1.PodList{}
	if err := r.Client.List(ctx, podList, client.InNamespace(namespace)); err != nil {
		return false, err
	}
	for _, p := range podList.Items {
		if p.UID == resource.UID {
			if p.Status.Phase == corev1.PodRunning {
				return true, nil
			}
			return false, nil
		}
	}
	return false, nil
}

func (r *KeptnWorkloadInstanceReconciler) isStatefulSetRunning(ctx context.Context, resource klcv1alpha1.ResourceReference, namespace string) (bool, error) {
	sts := appsv1.StatefulSet{}
	err := r.Client.Get(ctx, types.NamespacedName{Name: resource.Name, Namespace: namespace}, &sts)
	if err != nil {
		return false, err
	}
	return *sts.Spec.Replicas == sts.Status.AvailableReplicas, nil
}
