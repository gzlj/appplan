package workloadplan

const (
	WORKLOADTYPE_DEPLOYMENT = "DEPLOYMENT"
	WORKLOADTYPE_STATEFULSET = "STATEFULSET"
	WORKLOADTYPE_DAEMONSET = "DAEMONSET"

	CRON_PERIOD_YEAR = "YEAR"
	CRON_PERIOD_DAY = "DAY"
	CRON_PERIOD_MONTH = "MONTH"

	ANNOTATION_LAST_REPLICAS_KEY = "lastReplicas"

	DAEMONSET_NODE_SELECTOR_STOP = "daemonSetShouldStop"
)