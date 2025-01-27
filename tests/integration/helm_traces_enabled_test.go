package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SumoLogic/sumologic-kubernetes-collection/tests/integration/internal/ctxopts"
	"github.com/SumoLogic/sumologic-kubernetes-collection/tests/integration/internal/stepfuncs"
)

func Test_Helm_Traces_Enabled(t *testing.T) {
	const (
		tickDuration = time.Second
		waitDuration = time.Minute * 2
	)

	featInstall := features.New("installation").
		Assess("sumologic secret is created with endpoints",
			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
				terrak8s.WaitUntilSecretAvailable(t, ctxopts.KubectlOptions(ctx), "sumologic", 60, tickDuration)
				secret := terrak8s.GetSecret(t, ctxopts.KubectlOptions(ctx), "sumologic")
				// Apart from 10 default entries it also has 1 for traces
				require.Len(t, secret.Data, 11, "Secret has incorrect number of endpoints")
				return ctx
			}).
		Assess("fluentd logs statefulset is ready",
			stepfuncs.WaitUntilStatefulSetIsReady(
				waitDuration,
				tickDuration,
				stepfuncs.WithNameF(
					stepfuncs.ReleaseFormatter("%s-sumologic-fluentd-logs"),
				),
				stepfuncs.WithLabelsF(
					stepfuncs.LabelFormatterKV{
						K: "app",
						V: stepfuncs.ReleaseFormatter("%s-sumologic-fluentd-logs"),
					},
				),
			),
		).
		Assess("fluentd logs buffers PVCs are created",
			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
				namespace := ctxopts.Namespace(ctx)
				releaseName := ctxopts.HelmRelease(ctx)
				kubectlOptions := ctxopts.KubectlOptions(ctx)

				t.Logf("kubeconfig: %s", kubectlOptions.ConfigPath)
				cl, err := terrak8s.GetKubernetesClientFromOptionsE(t, kubectlOptions)
				require.NoError(t, err)

				assert.Eventually(t, func() bool {
					pvcs, err := cl.CoreV1().PersistentVolumeClaims(namespace).
						List(ctx, metav1.ListOptions{
							LabelSelector: fmt.Sprintf("app=%s-sumologic-fluentd-logs", releaseName),
						})
					if !assert.NoError(t, err) {
						return false
					}

					return err == nil && len(pvcs.Items) == 1
				}, waitDuration, tickDuration)
				return ctx
			}).
		Assess("fluentd metrics statefulset is ready",
			stepfuncs.WaitUntilStatefulSetIsReady(
				waitDuration,
				tickDuration,
				stepfuncs.WithNameF(
					stepfuncs.ReleaseFormatter("%s-sumologic-fluentd-metrics"),
				),
				stepfuncs.WithLabelsF(
					stepfuncs.LabelFormatterKV{
						K: "app",
						V: stepfuncs.ReleaseFormatter("%s-sumologic-fluentd-metrics"),
					},
				),
			),
		).
		Assess("fluentd metrics buffers PVCs are created",
			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
				namespace := ctxopts.Namespace(ctx)
				releaseName := ctxopts.HelmRelease(ctx)
				kubectlOptions := ctxopts.KubectlOptions(ctx)

				t.Logf("kubeconfig: %s", kubectlOptions.ConfigPath)
				cl, err := terrak8s.GetKubernetesClientFromOptionsE(t, kubectlOptions)
				require.NoError(t, err)

				assert.Eventually(t, func() bool {
					pvcs, err := cl.CoreV1().PersistentVolumeClaims(namespace).
						List(ctx, metav1.ListOptions{
							LabelSelector: fmt.Sprintf("app=%s-sumologic-fluentd-metrics", releaseName),
						})
					if !assert.NoError(t, err) {
						return false
					}

					return err == nil && len(pvcs.Items) == 1
				}, waitDuration, tickDuration)
				return ctx
			}).
		Assess("fluentd events statefulset is ready",
			stepfuncs.WaitUntilStatefulSetIsReady(
				waitDuration,
				tickDuration,
				stepfuncs.WithNameF(
					stepfuncs.ReleaseFormatter("%s-sumologic-fluentd-events"),
				),
				stepfuncs.WithLabelsF(
					stepfuncs.LabelFormatterKV{
						K: "app",
						V: stepfuncs.ReleaseFormatter("%s-sumologic-fluentd-events"),
					},
				),
			),
		).
		Assess("fluentd events buffers PVCs are created",
			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
				namespace := ctxopts.Namespace(ctx)
				releaseName := ctxopts.HelmRelease(ctx)
				kubectlOptions := ctxopts.KubectlOptions(ctx)

				t.Logf("kubeconfig: %s", kubectlOptions.ConfigPath)
				cl, err := terrak8s.GetKubernetesClientFromOptionsE(t, kubectlOptions)
				require.NoError(t, err)

				assert.Eventually(t, func() bool {
					pvcs, err := cl.CoreV1().PersistentVolumeClaims(namespace).
						List(ctx, metav1.ListOptions{
							LabelSelector: fmt.Sprintf("app=%s-sumologic-fluentd-events", releaseName),
						})
					if !assert.NoError(t, err) {
						return false
					}

					return err == nil && len(pvcs.Items) == 1
				}, waitDuration, tickDuration)
				return ctx
			}).
		Assess("prometheus pods are available",
			stepfuncs.WaitUntilPodsAvailable(
				metav1.ListOptions{
					LabelSelector: "app=prometheus",
				},
				1,
				waitDuration,
				tickDuration,
			),
		).
		Assess("fluent-bit daemonset is running",
			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
				var daemonsets []appsv1.DaemonSet
				require.Eventually(t, func() bool {
					daemonsets = terrak8s.ListDaemonSets(t, ctxopts.KubectlOptions(ctx), metav1.ListOptions{
						LabelSelector: "app.kubernetes.io/name=fluent-bit",
					})

					return len(daemonsets) == 1
				}, waitDuration, tickDuration)

				require.EqualValues(t, 0, daemonsets[0].Status.NumberUnavailable)
				return ctx
			}).
		Feature()

	featTraces := features.New("traces").
		// TODO: Rewrite into similar step func as WaitUntilStatefulSetIsReady but for deployments
		Assess("otelcol deployment is ready", func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			res := envConf.Client().Resources(ctxopts.Namespace(ctx))
			releaseName := ctxopts.HelmRelease(ctx)
			labelSelector := fmt.Sprintf("app=%s-sumologic-otelcol", releaseName)
			ds := appsv1.DeploymentList{}

			require.NoError(t,
				wait.For(
					conditions.New(res).
						ResourceListN(&ds, 1,
							resources.WithLabelSelector(labelSelector),
						),
					wait.WithTimeout(waitDuration),
					wait.WithInterval(tickDuration),
				),
			)
			require.NoError(t,
				wait.For(
					conditions.New(res).
						DeploymentConditionMatch(&ds.Items[0], appsv1.DeploymentAvailable, corev1.ConditionTrue),
					wait.WithTimeout(waitDuration),
					wait.WithInterval(tickDuration),
				),
			)
			return ctx
		}).
		// TODO: Rewrite into similar step func as WaitUntilStatefulSetIsReady but for daemonsets
		Assess("otelagent daemonset is ready", func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			res := envConf.Client().Resources(ctxopts.Namespace(ctx))
			nl := corev1.NodeList{}
			if !assert.NoError(t, res.List(ctx, &nl)) {
				return ctx
			}

			releaseName := ctxopts.HelmRelease(ctx)
			labelSelector := fmt.Sprintf("app=%s-sumologic-otelagent", releaseName)
			ds := appsv1.DaemonSetList{}

			require.NoError(t,
				wait.For(
					conditions.New(res).
						ResourceListN(&ds, 1,
							resources.WithLabelSelector(labelSelector),
						),
					wait.WithTimeout(waitDuration),
					wait.WithInterval(tickDuration),
				),
			)
			require.NoError(t,
				wait.For(
					conditions.New(res).
						ResourceMatch(&ds.Items[0], func(object k8s.Object) bool {
							d := object.(*appsv1.DaemonSet)
							return d.Status.NumberUnavailable == 0 &&
								d.Status.NumberReady == int32(len(nl.Items))
						}),
					wait.WithTimeout(waitDuration),
					wait.WithInterval(tickDuration),
				),
			)
			return ctx
		}).
		Feature()

	testenv.Test(t, featInstall, featTraces)
}
