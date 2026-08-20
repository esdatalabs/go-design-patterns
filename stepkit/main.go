package main

func main() {
	// cfg := config.Load()

	// gateways := make(map[string]query.ClusterGateway)
	// for _, c := range cfg.Clusters {
	// 	promClient := newPrometheusAPIClient(c.PrometheusURL)
	// 	gateways[c.Name] = query.NewPrometheusProvider(promClient)
	// }

	// notifiers := []notify.NotificationGateway{
	// 	notify.NewSlackProvider(cfg.SlackWebhookURL),
	// }

	// remediators := map[string]remediate.RemediationGateway{
	// 	"restart":  remediate.NewRestartProvider(cfg.K8sConfig),
	// 	"scale":    remediate.NewScaleProvider(cfg.K8sConfig),
	// 	"failover": remediate.NewFailoverProvider(cfg.K8sConfig),
	// }

	// p := pipeline.New(
	// 	discover.NewStep(discover.NewK8sProvider(cfg.K8sConfig)),
	// 	query.NewStep(gateways, cfg.KPIs, cfg.QueryTimeout, cfg.MaxConcurrency, logger),
	// 	emitmetrics.NewStep(promRegistry),
	// 	analyzekpi.NewStep(),
	// 	remediate.NewStep(remediators, cfg.CooldownPeriod),
	// 	notify.NewStep(notifiers),
	// )

	// runLoop(p, cfg.PollInterval)
}
