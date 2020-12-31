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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	cronv1 "github.com/believening/kubebuilder-example/cronjob/api/v1"
)

var _ = Describe("Cronjob controller", func() {
	const (
		CronjobName      = "test-cronjob"
		CronjobNamespace = "default"
		JobName          = "test-job"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("Updating Cronjob Status", func() {
		It("Should Increase Active Count", func() {
			By("Create a new cronjob")
			ctx := context.Background()
			cronJob := &cronv1.CronJob{
				TypeMeta: metav1.TypeMeta{
					Kind:       "CronJob",
					APIVersion: "examples.kubebuilder.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      CronjobName,
					Namespace: CronjobNamespace,
				},
				Spec: cronv1.CronJobSpec{
					Schedule: "1 * * * *",
					JobTemplate: batchv1beta1.JobTemplateSpec{
						Spec: batchv1.JobSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:            "test-container",
											Image:           "busybox",
											ImagePullPolicy: corev1.PullIfNotPresent,
										},
									},
									RestartPolicy: corev1.RestartPolicyOnFailure,
								},
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, cronJob)).Should(Succeed())

			cronjobLookupKey := types.NamespacedName{Name: CronjobName, Namespace: CronjobNamespace}
			createdCronjob := &cronv1.CronJob{}

			// `Eventually()` 重复执行直到超时或这满足期望
			Eventually(func() bool {
				err := k8sClient.Get(ctx, cronjobLookupKey, createdCronjob)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			Expect(createdCronjob.Spec.Schedule).Should(Equal("1 * * * *"))

			By("By checking the CronJob has zero active Jobs")
			//  `Consistently()` 在指定时间内重复执行且满足期望
			Consistently(func() (int, error) {
				err := k8sClient.Get(ctx, cronjobLookupKey, createdCronjob)
				if err != nil {
					return -1, err
				}
				return len(createdCronjob.Status.Active), nil
			}, duration, interval).Should(Equal(0))

			By("By creating a new Job")
			testJob := &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      JobName,
					Namespace: CronjobNamespace,
				},
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							// For simplicity, we only fill out the required fields.
							Containers: []corev1.Container{
								{
									Name:            "test-container",
									Image:           "busybox",
									ImagePullPolicy: corev1.PullIfNotPresent,
								},
							},
							RestartPolicy: corev1.RestartPolicyOnFailure,
						},
					},
				},
				Status: batchv1.JobStatus{
					Active: 2,
				},
			}

			gvk := cronv1.GroupVersion.WithKind("CronJob")

			controllerRef := metav1.NewControllerRef(createdCronjob, gvk)
			testJob.SetOwnerReferences([]metav1.OwnerReference{*controllerRef})
			Expect(k8sClient.Create(ctx, testJob)).Should(Succeed())

			By("By check job")
			jobLookupKey := types.NamespacedName{Name: JobName, Namespace: CronjobNamespace}
			createJob := &batchv1.Job{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, jobLookupKey, createJob)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			By("By checking that the CronJob has one active Job")
			Eventually(func() ([]string, error) {
				err := k8sClient.Get(ctx, cronjobLookupKey, createdCronjob)
				if err != nil {
					return nil, err
				}

				names := []string{}
				for _, job := range createdCronjob.Status.Active {
					names = append(names, job.Name)
				}
				return names, nil
			}, 2*timeout, interval).Should(ConsistOf(JobName), "should list our active job %s in the active jobs list in status", JobName)
		})
	})

})
