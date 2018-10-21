package gowrapmx4j

import (
	"sync"

	log "github.com/Sirupsen/logrus"
)

var registry = make(map[string]MX4JMetric)
var reglock = &sync.RWMutex{}

// Set a value in the Registry keyed to its Human Name
func RegistrySet(mm MX4JMetric, mb MX4JData) {
	reglock.Lock()
	defer reglock.Unlock()
	log.Debugf("RegistrySet: %s", mm.HumanName)

	mm.Data = mb
	registry[mm.HumanName] = mm
}

// Return a single MX4JMetric keyed by its human readable name
func RegistryGet(humanName string) MX4JMetric {
	reglock.RLock()
	defer reglock.RUnlock()

	return registry[humanName]
}

// Return all data points in the Registry
func RegistryBeans() map[string]MX4JData {
	reglock.RLock()
	defer reglock.RUnlock()

	beans := make(map[string]MX4JData)
	for hname, mm := range registry {
		beans[hname] = mm.Data
	}
	return beans
}

// Return a slice of all MX4JMetrics currently registered
func RegistryGetAll() []MX4JMetric {
	reglock.RLock()
	defer reglock.RUnlock()
	metrics := make([]MX4JMetric, 0, 0)
	for _, mm := range registry {
		metrics = append(metrics, mm)
	}
	return metrics
}

// Return a map of MX4JMetric structs keyed by their human readable name field.
func RegistryGetHRMap() map[string]MX4JMetric {
	reglock.RLock()
	defer reglock.RUnlock()

	metrics := make(map[string]MX4JMetric)
	for _, mm := range registry {
		metrics[mm.HumanName] = mm
	}
	return metrics
}

// Purge the gowrapmx4j data registry
// Primarily for use in the case where connection to MX4J has been lost,
// and reporting stale data is unhelpful. Endpoints will need to be re-registered
// in order for data collection to continue.
func RegistryPurge() {
	reglock.Lock()
	defer reglock.Unlock()

	// Replace the registry with a new map
	registry = make(map[string]MX4JMetric)
}

// RegistryFlush resets the MX4JMetric.Data fields for all registered metrics.
// If connection to MX4J is lost this can be called to remove stale data but keep the
// metric handles for when MX4J recovers.
func RegistryFlush() {
	reglock.Lock()
	defer reglock.Unlock()

	// Replace the registry with a new map
	for k, mm := range registry {
		log.Debugf("Blanking gowrapmx4j.registry data of %s", k)
		mm.Data = nil
		registry[mm.HumanName] = mm
	}
}
