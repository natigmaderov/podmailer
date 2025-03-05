/*
Copyright 2025.

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

package controller

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	podmailerv1alpha1 "github.com/natigmaderov/podmailer/api/v1alpha1"
	"github.com/natigmaderov/podmailer/internal/mail"
)

// PodMailerReconciler reconciles a PodMailer object
type PodMailerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=podmailer.podmailer.io,resources=podmailers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=podmailer.podmailer.io,resources=podmailers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=podmailer.podmailer.io,resources=podmailers/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PodMailer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/reconcile
func (r *PodMailerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the PodMailer instance
	podMailer := &podmailerv1alpha1.PodMailer{}
	if err := r.Get(ctx, req.NamespacedName, podMailer); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// List pods in the specified namespaces or all namespaces
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{}

	if len(podMailer.Spec.Namespaces) > 0 {
		// Monitor specific namespaces
		var downPods []podmailerv1alpha1.PodStatus
		for _, ns := range podMailer.Spec.Namespaces {
			pods := &corev1.PodList{}
			if err := r.List(ctx, pods, client.InNamespace(ns)); err != nil {
				log.Error(err, "Failed to list pods in namespace", "namespace", ns)
				continue
			}
			downPods = append(downPods, checkPodsStatus(pods.Items)...)
		}
		podMailer.Status.DownPods = downPods
	} else {
		// Monitor all namespaces
		if err := r.List(ctx, podList, listOpts...); err != nil {
			log.Error(err, "Failed to list pods")
			return ctrl.Result{}, err
		}
		podMailer.Status.DownPods = checkPodsStatus(podList.Items)
	}

	// Update LastCheckTime
	now := metav1.Now()
	podMailer.Status.LastCheckTime = &now

	// Send notifications if there are down pods
	if len(podMailer.Status.DownPods) > 0 {
		mailer := mail.NewMailer(podMailer.Spec.SMTP)
		if err := mailer.SendPodDownNotification(podMailer.Spec.Recipients, podMailer.Status.DownPods); err != nil {
			log.Error(err, "Failed to send email notification")
			return ctrl.Result{}, err
		}
		podMailer.Status.LastNotificationTime = &now
	}

	// Update status
	if err := r.Status().Update(ctx, podMailer); err != nil {
		log.Error(err, "Failed to update PodMailer status")
		return ctrl.Result{}, err
	}

	// Requeue based on CheckInterval
	return ctrl.Result{
		RequeueAfter: time.Duration(podMailer.Spec.CheckInterval) * time.Second,
	}, nil
}

func checkPodsStatus(pods []corev1.Pod) []podmailerv1alpha1.PodStatus {
	var downPods []podmailerv1alpha1.PodStatus

	for _, pod := range pods {
		if isPodDown(&pod) {
			downPods = append(downPods, podmailerv1alpha1.PodStatus{
				Name:      pod.Name,
				Namespace: pod.Namespace,
				Status:    string(pod.Status.Phase),
			})
		}
	}

	return downPods
}

func isPodDown(pod *corev1.Pod) bool {
	if pod.Status.Phase == corev1.PodFailed || pod.Status.Phase == corev1.PodUnknown {
		return true
	}

	if pod.Status.Phase == corev1.PodPending {
		// Check if pod is stuck in pending state
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodScheduled &&
				condition.Status == corev1.ConditionFalse &&
				time.Since(condition.LastTransitionTime.Time).Minutes() > 5 {
				return true
			}
		}
	}

	return false
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodMailerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&podmailerv1alpha1.PodMailer{}).
		Watches(
			&corev1.Pod{},
			handler.EnqueueRequestsFromMapFunc(r.findPodMailersForPod),
		).
		Complete(r)
}

// findPodMailersForPod maps a Pod to a list of PodMailer objects that should be reconciled
func (r *PodMailerReconciler) findPodMailersForPod(ctx context.Context, obj client.Object) []reconcile.Request {
	var requests []reconcile.Request

	// Get all PodMailer instances
	podMailerList := &podmailerv1alpha1.PodMailerList{}
	if err := r.List(ctx, podMailerList); err != nil {
		return nil
	}

	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return nil
	}

	for _, pm := range podMailerList.Items {
		// Check if this pod's namespace is being monitored by the PodMailer
		if len(pm.Spec.Namespaces) == 0 || contains(pm.Spec.Namespaces, pod.Namespace) {
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      pm.Name,
					Namespace: pm.Namespace,
				},
			})
		}
	}

	return requests
}

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
