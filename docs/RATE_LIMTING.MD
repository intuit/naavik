# Rate Limting Using TrafficConfig

TrafficConfig spec can be used to configure Rate Limiting on the service by setting the QuotaGroup

Rate Limiting can be configured on the service by using the QuotaGroup Fields of the [TrafficConfig](https://github.com/istio-ecosystem/admiral-api/blob/master/pkg/apis/admiral/v1/traffic_config.go) spec.

QuotaGroups are of two types
1. Total Quota - Applies to all incoming clients/apps.
2. App Quota - Applies to the assoicated clients/apps.

Envoy supports two flavors of rate limiting
1. Global - Uses an external database(redis) for storing the counters.
2. Local - Counters are stored locally within each replica.

Naavik currently supports only Local Rate Limiting for both Total and App Quotas.

The `MaxAmount` set in the rate limiting spec is set per replica. So if 100 TPS is desired for the entire service which is running with 5 replicas, then the `MaxValue` should be set to 20.

The App Rate Limiting based on associated apps relies on the header with the name set with startup param `traffic_config_identity_key` to be present in the request. The quota is unique for each associated app.

### TODO
1. Accept `MaxAmount` for the entire service and dynamically determine the quota for each replica.
2. Add support for Global Rate Limiting.