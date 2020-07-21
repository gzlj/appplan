package workloadplan

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

func SetDeploymentZeroReplicas(deploy *appsv1.Deployment) (err error) {
	if deploy == nil {
		return
	}
	var size int32 = 0
	deploy.Spec.Replicas = &size
	err = K8sClient.Update(context.TODO(), deploy)
	if err != nil {
		log.Info("Failed to scale for deployment %s/%s", deploy.Namespace, deploy.Name)
	}
	return
}

func ScaleUpDeploymentReplicas(deploy *appsv1.Deployment) (err error) {
	if deploy == nil || *deploy.Spec.Replicas != 0{
		return
	}
	var size int32 = 1
	sizeStr, ok := deploy.Annotations[ANNOTATION_LAST_REPLICAS_KEY]
	if ok {
		lastReplicas, err := strconv.ParseInt(sizeStr, 10, 32)
		//lastReplicas, err := strconv.Atoi(sizeStr)
		if err == nil {
			size  = int32(lastReplicas)
		}
	}
	deploy.Spec.Replicas = &size
	err = K8sClient.Update(context.TODO(), deploy)
	if err != nil {
		log.Info("Failed to scale for deployment %s/%s", deploy.Namespace, deploy.Name)
	}
	return
}

func SetStatefulSetZeroReplicas(deploy *appsv1.StatefulSet) (err error) {
	if deploy == nil {
		return
	}
	var size int32 = 0
	deploy.Spec.Replicas = &size
	err = K8sClient.Update(context.TODO(), deploy)
	if err != nil {
		log.Info("Failed to scale for statefulset %s/%s", deploy.Namespace, deploy.Name)
	}
	return
}

func ScaleUpStatefulSetZeroReplicas(deploy *appsv1.StatefulSet) (err error) {
	if deploy == nil || *deploy.Spec.Replicas != 0{
		return
	}
	var size int32 = 1
	sizeStr, ok := deploy.Annotations[ANNOTATION_LAST_REPLICAS_KEY]
	if ok {
		lastReplicas, err := strconv.ParseInt(sizeStr, 10, 32)
		//lastReplicas, err := strconv.Atoi(sizeStr)
		if err == nil {
			size  = int32(lastReplicas)
		}
	}
	deploy.Spec.Replicas = &size
	err = K8sClient.Update(context.TODO(), deploy)
	if err != nil {
		log.Info("Failed to scale for deployment %s/%s", deploy.Namespace, deploy.Name)
	}
	return
}

func SetDaemonSetZeroReplicas(deploy *appsv1.DaemonSet) (err error) {
	if deploy == nil {
		return
	}
	if deploy.Spec.Template.Spec.NodeSelector == nil {
		deploy.Spec.Template.Spec.NodeSelector = make(map[string]string)
	}
	deploy.Spec.Template.Spec.NodeSelector[DAEMONSET_NODE_SELECTOR_STOP] = "true"
	err = K8sClient.Update(context.TODO(), deploy)
	if err != nil {
		log.Info("Failed to update daemonset %s/%s", deploy.Namespace, deploy.Name)
	}
	return
}

func ScaleUpDaemonSetReplicas(deploy *appsv1.DaemonSet) (err error) {
	if deploy == nil {
		return
	}
	nodeSelector := deploy.Spec.Template.Spec.NodeSelector
	if nodeSelector == nil {
		return
	}
	delete(nodeSelector, DAEMONSET_NODE_SELECTOR_STOP)
	err = K8sClient.Update(context.TODO(), deploy)
	if err != nil {
		log.Info("Failed to scale up daemonset %s/%s", deploy.Namespace, deploy.Name)
	}
	return
}

func GenerateHelmWorkloadLabels(release string) map[string]string {
	labels := map[string]string{
		LABEL_KEY_HELM: LABEL_VALUE_HELM,
		LABEL_KEY_HELM_RELEASE : release,
	}
	return labels
}

func SetHelmWorkloadReplicasZero(namespace, release string, scaleUp bool) (errs []error){
	//podList := &corev1.PodList{}
	//err = cl.List(context.Background(), podList, client.InNamespace("default"))
	/*
		labels := map[string]string{"test-label": "test-pod-2"}
		err := informerCache.List(context.Background(), &out, client.MatchingLabels(labels))
	*/
	var err error
	labels := GenerateHelmWorkloadLabels(release)
	deployList := &appsv1.DeploymentList{}
	err = K8sClient.List(context.TODO(), deployList, client.InNamespace(namespace), client.MatchingLabels(labels))
	if err == nil {
		for _, deploy := range deployList.Items {
			if scaleUp {
				ScaleUpDeploymentReplicas(&deploy)
			} else {
				SetDeploymentZeroReplicas(&deploy)
			}

		}

	} else {
		errs = append(errs, err)
	}
	statefulSetList := &appsv1.StatefulSetList{}
	err = K8sClient.List(context.TODO(), statefulSetList, client.InNamespace(namespace), client.MatchingLabels(labels))
	if err == nil {
		for _, deploy := range statefulSetList.Items {
			if scaleUp {
				ScaleUpStatefulSetZeroReplicas(&deploy)
			} else {
				SetStatefulSetZeroReplicas(&deploy)
			}

		}
	} else {
		errs = append(errs, err)
	}
	dstList := &appsv1.DaemonSetList{}
	err = K8sClient.List(context.TODO(), dstList, client.InNamespace(namespace), client.MatchingLabels(labels))
	if err == nil {
		for _, deploy := range dstList.Items {
			if scaleUp {
				ScaleUpDaemonSetReplicas(&deploy)
			} else {
				SetDaemonSetZeroReplicas(&deploy)
			}
		}
	} else {
		errs = append(errs, err)
	}
	return
}