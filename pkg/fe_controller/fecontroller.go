package fe_controller

import (
	"context"
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/api/v1alpha1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type FeController struct {
	k8sclient client.Client
	k8sReader client.Reader
}

func New(k8sclient client.Client, k8sReader client.Reader) *FeController {
	return &FeController{
		k8sclient: k8sclient,
		k8sReader: k8sReader,
	}
}

//Sync starRocksCluster spec to fe statefulset and service.
func (fc *FeController) Sync(ctx context.Context, src *srapi.StarRocksCluster) error {
	if src.Spec.StarRocksFeSpec == nil {
		klog.Info("FeController Sync", "the fe component is not needed", "namespace", src.Namespace, "starrocks cluster name", src.Name)
		return nil
	}

	//generate new fe service.
	svc := rutils.BuildService(src, rutils.FeService)
	fs := srapi.StarRocksFeStatus{ServiceName: svc.Name}
	src.Status.StarRocksFeStatus = &fs
	//create or update fe service, update the status of fe on src.
	if err := k8sutils.CreateOrUpdateService(ctx, fc.k8sclient, &svc); err != nil {
		klog.Error("FeController Sync", "create or update service namespace", svc.Namespace, "name", svc.Name)
		return err
	}
	feFinalizers := []string{srapi.FE_SERVICE_FINALIZER}
	defer func() {
		rutils.MergeSlices(src.Finalizers, feFinalizers)
	}()

	//create fe statefulset.
	st := rutils.NewStatefulset(fc.buildStatefulSetParams(src))
	var est appv1.StatefulSet
	err := fc.k8sclient.Get(ctx, types.NamespacedName{Namespace: st.Namespace, Name: st.Name}, &est)
	if err != nil && apierrors.IsNotFound(err) {
		fs.ResourceNames = append(fs.ResourceNames, st.Name)
		feFinalizers = append(feFinalizers, srapi.FE_STATEFULSET_FINALIZER)
		return k8sutils.CreateClientObject(ctx, fc.k8sclient, &st)
	} else if err != nil {
		return err
	}

	//if the spec is not change, update the status of fe on src.
	if rutils.StatefulSetDeepEqual(&st, est) {
		fs.ResourceNames = []string{st.Name}
		if err := fc.UpdateFeStatus(&fs, st); err != nil {
			return err
		}
		//no update
		return nil
	}

	//fe spec changed update the statefulset.
	rutils.MergeStatefulSets(&st, est)
	if err := k8sutils.UpdateClientObject(ctx, fc.k8sclient, &st); err != nil {
		return err
	}

	return fc.UpdateFeStatus(&fs, st)
}

//buildStatefulSetParams generate the params of construct the statefulset.
func (fc *FeController) buildStatefulSetParams(src *srapi.StarRocksCluster) rutils.StatefulSetParams {
	feSpec := src.Spec.StarRocksFeSpec
	var pvcs []corev1.PersistentVolumeClaim
	for _, vm := range feSpec.StorageVolumes {
		pvcs = append(pvcs, corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{Name: vm.Name},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				StorageClassName: vm.StorageClassName,
			},
		})
	}

	var spname string
	spname = src.Name + "-" + srapi.DEFAULT_FE
	if feSpec.Name != "" {
		spname = feSpec.Name
	}
	var labels rutils.Labels
	labels[srapi.OwnerReference] = src.Name
	labels[srapi.ComponentLabelKey] = srapi.DEFAULT_FE
	labels.AddLabel(src.Labels)
	or := metav1.OwnerReference{
		UID:        src.UID,
		Kind:       src.Kind,
		APIVersion: src.APIVersion,
		Name:       src.Name,
	}

	return rutils.StatefulSetParams{
		Name:                 spname,
		Namespace:            src.Namespace,
		VolumeClaimTemplates: pvcs,
		ServiceName:          srapi.DEFAULT_FE_SERVICE_NAME,
		PodTemplateSpec:      fc.buildPodTemplate(src),
		Labels:               labels,
		Selector:             labels,
		OwnerReferences:      []metav1.OwnerReference{or},
	}
}

//UpdateFeStatus update the starrockscluster fe status.
func (fc *FeController) UpdateFeStatus(fs *srapi.StarRocksFeStatus, st appv1.StatefulSet) error {
	var podList corev1.PodList
	if err := fc.k8sReader.List(context.Background(), &podList, client.InNamespace(st.Namespace), client.MatchingLabels(st.Spec.Selector.MatchLabels)); err != nil {
		return err
	}

	var creatings, runnings, faileds []string
	var podmap map[string]corev1.Pod
	//get all pod status that controlled by st.
	for _, pod := range podList.Items {
		//TODO: test
		podmap[pod.Name] = pod
		if pod.Status.Phase == corev1.PodPending {
			creatings = append(creatings, pod.Name)
		} else if pod.Status.Phase == corev1.PodRunning {
			runnings = append(runnings, pod.Name)
		} else {
			faileds = append(faileds, pod.Name)
		}
	}

	fs.Phase = srapi.ComponentReconciling
	if len(runnings) == int(*st.Spec.Replicas) {
		fs.Phase = srapi.ComponentRunning
	} else if len(faileds) != 0 {
		fs.Phase = srapi.ComponentFailed
		fs.Reason = podmap[faileds[0]].Status.Reason
	} else if len(creatings) != 0 {
		fs.Phase = srapi.ComponentPending
		fs.Reason = podmap[creatings[0]].Status.Reason
	}

	return nil
}

//buildPodTemplate construct the podTemplate for deploy fe.
func (fc *FeController) buildPodTemplate(src *srapi.StarRocksCluster) corev1.PodTemplateSpec {
	metaname := src.Name + "-fe"
	labels := src.Labels
	feSpec := src.Spec.StarRocksFeSpec
	if feSpec.Name != "" {
		labels[srapi.OwnerReference] = feSpec.Name
	} else {
		labels[srapi.OwnerReference] = srapi.DEFAULT_FE
	}

	vols := []corev1.Volume{
		{
			Name: srapi.DEFAULT_FE_CONFIG_NAME,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: srapi.DEFAULT_FE_CONFIG_NAME,
					},
				},
			},
		},
	}

	var volMounts []corev1.VolumeMount
	for _, vm := range feSpec.StorageVolumes {
		volMounts = append(volMounts, corev1.VolumeMount{
			Name:      vm.Name,
			MountPath: vm.MountPath,
		})
	}

	operatorContainers := []corev1.Container{
		{
			Name:  srapi.DEFAULT_FE,
			Image: feSpec.Image,
			//TODO: 增加启动
			Command: []string{},
			//TODO: 使用args
			Args: []string{},
			Ports: []corev1.ContainerPort{{
				Name:          "http_port",
				ContainerPort: 8030,
				Protocol:      corev1.ProtocolTCP,
			}, {
				Name:          "rpc_port",
				ContainerPort: 9020,
				Protocol:      corev1.ProtocolTCP,
			}, {
				Name:          "query_port",
				ContainerPort: 9030,
				Protocol:      corev1.ProtocolTCP,
			},
			},
			Env: []corev1.EnvVar{
				{
					Name: "POD_NAME",
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					},
				}, {
					Name: "POD_NAMESPACE",
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"},
					},
				},
			},
			Resources:    feSpec.ResourceRequirements,
			VolumeMounts: volMounts,
			//TODO: LivenessProbe,ReadinessProbe
			ImagePullPolicy: corev1.PullIfNotPresent,
		},
	}

	podSpec := corev1.PodSpec{
		Volumes:            vols,
		Containers:         operatorContainers,
		ServiceAccountName: src.Spec.ServiceAccount,
	}

	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:        metaname,
			Namespace:   src.Namespace,
			Labels:      labels,
			Annotations: src.Annotations,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: src.APIVersion,
					Kind:       src.Kind,
					Name:       src.Name,
				},
			},
		},
		Spec: podSpec,
	}
}
