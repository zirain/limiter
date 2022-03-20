/*
Copyright 2022.

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
	"reflect"

	istioclient "istio.io/client-go/pkg/clientset/versioned"
	istioinformers "istio.io/client-go/pkg/informers/externalversions"
	"istio.io/client-go/pkg/listers/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	policyv1alpha1 "github.com/zirain/limiter/api/v1alpha1"
	"github.com/zirain/limiter/pkg/envoyfilter"
)

// RateLimitReconciler reconciles a RateLimit object
type RateLimitReconciler struct {
	client.Client
	istioClient istioclient.Interface
	Scheme      *runtime.Scheme

	envoyfilterInformer cache.SharedIndexInformer
	envoyfilterLister   v1alpha3.EnvoyFilterLister
}

func NewRateLimitReconciler(client client.Client, istioClient istioclient.Interface, istioInformerFactory istioinformers.SharedInformerFactory, scheme *runtime.Scheme) *RateLimitReconciler {
	efInformer := istioInformerFactory.Networking().V1alpha3().EnvoyFilters().Informer()
	efLister := istioInformerFactory.Networking().V1alpha3().EnvoyFilters().Lister()

	return &RateLimitReconciler{
		Client:              client,
		istioClient:         istioClient,
		Scheme:              scheme,
		envoyfilterLister:   efLister,
		envoyfilterInformer: efInformer,
	}
}

//+kubebuilder:rbac:groups=networking.istio.io,resources=envoyfilters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=policy.zirain.info,resources=ratelimits,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=policy.zirain.info,resources=ratelimits/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=policy.zirain.info,resources=ratelimits/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RateLimit object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *RateLimitReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	controllerLog := log.FromContext(ctx)

	var ratelimitList policyv1alpha1.RateLimitList
	if err := r.List(ctx, &ratelimitList, client.InNamespace(req.Namespace)); err != nil {
		controllerLog.Error(err, "unable to list child Jobs")
		return ctrl.Result{}, err
	}

	for _, rl := range ratelimitList.Items {
		generated := envoyfilter.ToEnvoyFilter(&rl)

		ef, err := r.envoyfilterLister.EnvoyFilters(generated.Namespace).Get(generated.Name)
		if err != nil {
			if errors.IsNotFound(err) {
				// create
				_, err := r.istioClient.NetworkingV1alpha3().EnvoyFilters(generated.Namespace).Create(context.TODO(), generated, metav1.CreateOptions{})
				if err != nil {
					return ctrl.Result{}, err
				}
			}
		} else {
			// update
			if !reflect.DeepEqual(generated.Spec, ef.Spec) {
				copiedEf := ef.DeepCopy()
				copiedEf.Spec = generated.Spec
				_, err := r.istioClient.NetworkingV1alpha3().EnvoyFilters(ef.Namespace).Update(context.TODO(), copiedEf, metav1.UpdateOptions{})
				if err != nil {
					return ctrl.Result{}, err
				}
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RateLimitReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&policyv1alpha1.RateLimit{}).
		Complete(r)
}
