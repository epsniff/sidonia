package cassandra

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	log "github.com/Sirupsen/logrus"
	"github.com/lytics/gowrapmx4j"
)

// Http response function extracts the state of the current node given the
// hostname parameter.
// Requires that "NodeStatus" metric be registered as below and in the cassandra
// entry point example.
/*
  mm := gowrapmx4j.MX4JMetric{HumanName: "NodeStatus", ObjectName: "org.apache.cassandra.net:type=FailureDetector",
		ValFunc: gowrapmx4j.DistillAttributeTypes}
	gowrapmx4j.RegistrySet(mm, nil)
*/
func HttpNodeStatus(hostname string) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		mjs := make(map[string]interface{})
		nsb := gowrapmx4j.RegistryGet("NodeStatus")
		if nsb.Data == nil {
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, fmt.Sprintf("NodeStatus Metric Data is nil: %#v", nsb), 500)
			return
		}
		metricMap, err := gowrapmx4j.DistillAttributeTypes(nsb.Data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, fmt.Sprintf("nodeStatus: Error extracting node status data: %v", err), 500)
			return
		}

		states, ok := metricMap["SimpleStates"]
		log.Debugf("%#v", states)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, fmt.Sprintf("nodeStatus: Error extracting node status data;  Key: SimpleStates not in data map"), 500)
			return
		}
		ss := states.(map[string]interface{})

		var hostKey string
		hostMatch := regexp.MustCompile(fmt.Sprintf(".*%s.*", hostname))
		for k, v := range ss {
			log.Debug("Nodestatus: %s %#v", k, v)
			if hostMatch.MatchString(k) {
				hostKey = k
				break
			}
		}

		mjs[hostKey], ok = ss[hostKey]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, fmt.Sprintf("nodeStatus: Error finding hostname=\"%s\" as key in list of nodes: %#v", hostname, err), 500)
			return
		}

		js, err := json.Marshal(mjs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, fmt.Sprintf("nodeStatus: Error marshaling JSON from MX4J data: %#v", err), 500)
			return
		}
		fmt.Fprintf(w, "%s", js)
	}
}
