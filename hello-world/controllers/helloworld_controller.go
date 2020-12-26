/*


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

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	examplesv1 "github.com/believening/kubebuilder-example/hello-world/api/v1"
)

// HelloWorldReconciler reconciles a HelloWorld object
type HelloWorldReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=examples.kubebuilder.io,resources=helloworldren,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=examples.kubebuilder.io,resources=helloworldren/status,verbs=get;update;patch

func (r *HelloWorldReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("helloworld", req.NamespacedName)

	// your logic here
	obj := new(examplesv1.HelloWorld)
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Deleted HelloWorld")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Get Failed", "helloworld", req.NamespacedName)
		return ctrl.Result{}, err
	}

	log.Info("Get HelloWorld", "Greet", obj.Spec.Greet)

	// return ctrl.Result{RequeueAfter: time.Second}, nil
	return ctrl.Result{}, nil
}

func (r *HelloWorldReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&examplesv1.HelloWorld{}).
		Complete(r)
}
