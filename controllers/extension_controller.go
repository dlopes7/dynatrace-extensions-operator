/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"hash/fnv"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strconv"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dynatracev1alpha1 "quay.io/dlopes7/dt-extensions-operator/api/v1alpha1"
)

const AnnotationTemplateHash = "internal.extension.dynatrace.com/template-hash"

// ExtensionReconciler reconciles a Extension object
type ExtensionReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=dynatrace.com,resources=extensions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dynatrace.com,resources=extensions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dynatrace.com,resources=extensions/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete

// Reconcile makes sure that the extensions are deployed for nodes where the OneAgent is running
func (r *ExtensionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("extension", req.NamespacedName)

	r.Log.Info("Reconciling extensions")

	instance := &dynatracev1alpha1.Extension{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	r.Log.Info("Found instance", "instance", instance)

	dsDesired, err := newDaemonSet(instance)
	if err != nil {
		r.Log.Info("Failed to get desired daemonset")
		return reconcile.Result{}, err
	}

	dsActual := &appsv1.DaemonSet{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: dsDesired.Name, Namespace: dsDesired.Namespace}, dsActual)
	if err != nil && k8serrors.IsNotFound(err) {
		r.Log.Info("Creating new daemonset")
		if err = r.Client.Create(ctx, dsDesired); err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	} else if hasDaemonSetChanged(dsDesired, dsActual) {
		r.Log.Info("Updating existing daemonset")
		if err = r.Client.Update(ctx, dsDesired); err != nil {
			return reconcile.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func getTemplateHash(a metav1.Object) string {
	if annotations := a.GetAnnotations(); annotations != nil {
		return annotations[AnnotationTemplateHash]
	}
	return ""
}

func hasDaemonSetChanged(a, b *appsv1.DaemonSet) bool {
	return getTemplateHash(a) != getTemplateHash(b)
}

func newPodSpec() corev1.PodSpec {

	p := corev1.PodSpec{
		Volumes: []corev1.Volume{
			{
				Name: "host-root",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/",
					},
				},
			},
		},
		Containers: []corev1.Container{{
			Image:           "quay.io/dlopes7/dt-extension:latest",
			ImagePullPolicy: corev1.PullAlways,
			Name:            "dt-extension",
			ReadinessProbe: &corev1.Probe{
				Handler: corev1.Handler{
					Exec: &corev1.ExecAction{
						Command: []string{
							"/extension-watcher", "health-check",
						},
					},
				},
				InitialDelaySeconds: 30,
				PeriodSeconds:       30,
				TimeoutSeconds:      1,
			},
			// SecurityContext: secCtx,
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "host-root",
					MountPath: "/mnt/root",
				},
			}},
		},
	}

	return p
}

func newDaemonSet(instance *dynatracev1alpha1.Extension) (*appsv1.DaemonSet, error) {

	podSpec := newPodSpec()

	selectorLabels := buildLabels(instance.GetName())
	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        instance.GetName(),
			Labels:      selectorLabels,
			Namespace:   instance.GetNamespace(),
			Annotations: map[string]string{},
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{MatchLabels: selectorLabels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: selectorLabels,
				},
				Spec: podSpec,
			},
		},
	}
	dsHash, err := generateDaemonSetHash(ds)
	if err != nil {
		return nil, err
	}
	ds.Annotations[AnnotationTemplateHash] = dsHash

	return ds, nil
}

func buildLabels(name string) map[string]string {
	return map[string]string{
		"dynatrace.com/component":          "extension",
		"extension.dynatrace.com/instance": name,
	}
}
func generateDaemonSetHash(ds *appsv1.DaemonSet) (string, error) {
	data, err := json.Marshal(ds)
	if err != nil {
		return "", err
	}

	hasher := fnv.New32()
	_, err = hasher.Write(data)
	if err != nil {
		return "", err
	}

	return strconv.FormatUint(uint64(hasher.Sum32()), 10), nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ExtensionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dynatracev1alpha1.Extension{}).
		Owns(&appsv1.DaemonSet{}).
		Complete(r)
}
