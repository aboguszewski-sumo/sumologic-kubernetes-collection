package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/SumoLogic/sumologic-kubernetes-collection/tests/integration/internal"
	"github.com/SumoLogic/sumologic-kubernetes-collection/tests/integration/internal/ctxopts"
	"github.com/SumoLogic/sumologic-kubernetes-collection/tests/integration/internal/stepfuncs"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func Test_Helm_Otelcol_Traces(t *testing.T) {
	const (
		tickDuration         = 3 * time.Second
		waitDuration         = 3 * time.Minute
		tracesCount     uint = 10 // number of traces generated
		spansPerTrace   uint = 5
		totalSpansCount uint = tracesCount * spansPerTrace
	)
	featInstall := features.New("traces").
		Assess("sumologic secret is created with endpoints",
			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
				terrak8s.WaitUntilSecretAvailable(t, ctxopts.KubectlOptions(ctx), "sumologic", 60, tickDuration)
				secret := terrak8s.GetSecret(t, ctxopts.KubectlOptions(ctx), "sumologic")
				require.Len(t, secret.Data, 2, "Secret has incorrect number of endpoints. There should be only 1 endpoint, for logs.")
				return ctx
			}).
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
		}).Feature()

	featTraces := features.New("traces").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			// TODO: This should be refactored. See internal/stepfuncs/logs.go for inspirations.
			release := ctxopts.HelmRelease(ctx)

			opts := ctxopts.KubectlOptions(ctx)

			colName := fmt.Sprintf("%s-sumologic-otelcol", release)
			args := []string{
				"run", "cst",
				"--image", "localhost:5001/kubernetes-tools:dev-latest",
				fmt.Sprintf("--serviceaccount=%s-sumologic", release),
				"--env", fmt.Sprintf("TOTAL_TRACES=%d", tracesCount),
				"--env", fmt.Sprintf("SPANS_PER_TRACE=%d", spansPerTrace),
				"--env", fmt.Sprintf("COLLECTOR_HOSTNAME=%s", colName),
				"--",
				"customer-trace-tester",
			}
			terrak8s.RunKubectl(t, opts, args...)
			return ctx
		}).Assess("wait for spans", stepfuncs.WaitUntilExpectedSpansPresent(
		spansPerTrace*tracesCount*4, // The generator sends spans from four sources
		map[string]string{},
		internal.ReceiverMockNamespace,
		internal.ReceiverMockServiceName,
		internal.ReceiverMockServicePort,
		waitDuration,
		tickDuration,
	)).Feature()

	testenv.Test(t, featInstall, featTraces)
}
