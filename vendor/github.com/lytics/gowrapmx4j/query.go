package gowrapmx4j

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/rcrowley/go-metrics"
)

// Query all registered MX4J endpoints and compose their data into the MX4JMetric
// array or return error
func QueryMX4J(mx4j MX4JService) (*[]MX4JMetric, error) {
	reg := RegistryGetAll()

	for _, mm := range reg {
		var newData MX4JData
		var err error
		data := mm.Data
		log.Debugf("Metric being queried: %#v", mm)

		// If first time querying endpoint, create data struct
		if data == nil {
			newData = Bean{}
			mx4jData, err := newData.QueryMX4J(mx4j, mm)
			if err != nil {
				retErr := fmt.Errorf("QueryMX4J Error: %v%s", newData, err)
				return nil, retErr
			}
			RegistrySet(mm, mx4jData)
		} else {
			newData, err = data.QueryMX4J(mx4j, mm)
			RegistrySet(mm, newData)
		}

		if mm.MetricFunc != nil && newData != nil {
			log.Debugf("Metric func running: %s", mm.HumanName)
			mm.MetricFunc(mm.Data, mm.HumanName)
		}

		if newData == nil {
			log.Errorf("No data returned from querying; blanking the metric registries")
			metrics.DefaultRegistry.UnregisterAll()
			RegistryFlush()
		}

		if err != nil {
			retErr := fmt.Errorf("QueryMX4J Error: %v %s", newData, err)
			return nil, retErr
		}
	}

	updated := RegistryGetAll()
	return &updated, nil
}
