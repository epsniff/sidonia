package gowrapmx4j

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

// Cassandra MX4J status endpoint
func HttpRegistryRaw(w http.ResponseWriter, r *http.Request) {
	mbeans := RegistryBeans()
	js, err := json.Marshal(mbeans)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "HttpRegistryRaw: Error marshaling JSON from MX4J data: %v", err)
	}
	fmt.Fprintf(w, "%s", js)
}

// API Endpoint which will execute the optionally specified ValFunc function
// on the data structure to process the metric's data.
func HttpRegistryProcessed(w http.ResponseWriter, r *http.Request) {
	metrics := RegistryGetAll()

	mjs := make(map[string]interface{})
	for _, m := range metrics {
		if m.ValFunc != nil {
			log.Infof("%s", m.HumanName)
			mdata, err := m.ValFunc(m.Data)
			if err != nil {
				log.Errorf("Error running value function for %s: %v", m.HumanName, err)
				continue
			}
			mjs[m.HumanName] = mdata
		} else {
			mjs[m.HumanName] = m.Data
		}
	}

	js, err := json.Marshal(mjs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("HttpRegistryProcessed: Error marshaling JSON from MX4J data: %#v", err), 500)
	}
	fmt.Fprintf(w, "%s", js)
}
