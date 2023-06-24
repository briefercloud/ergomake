package servicelogs

import "time"

func getLogsQuery(serviceID string, namespace string, container string, offset int64) map[string]interface{} {
	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []interface{}{
					map[string]interface{}{
						"term": map[string]interface{}{
							"kubernetes.labels.preview_ergomake_dev/id": serviceID,
						},
					},
					map[string]interface{}{
						"term": map[string]interface{}{
							"kubernetes.namespace": namespace,
						},
					},
					map[string]interface{}{
						"term": map[string]interface{}{
							"container.id": container,
						},
					},
					map[string]interface{}{
						"range": map[string]interface{}{
							"log.offset": map[string]interface{}{"gt": offset},
						},
					},
				},
			},
		},
		"sort": []map[string]interface{}{{
			"log.offset": map[string]interface{}{
				"order": "asc",
			},
		}},
		"size": 100,
	}
}

func getNextContainerQuery(serviceID string, namespace string, timestamp *time.Time) map[string]interface{} {
	filter := []interface{}{
		map[string]interface{}{
			"term": map[string]interface{}{
				"kubernetes.labels.preview_ergomake_dev/id": serviceID,
			},
		},
		map[string]interface{}{
			"term": map[string]interface{}{
				"kubernetes.namespace": namespace,
			},
		},
	}

	if timestamp != nil {
		filter = append(filter, map[string]interface{}{
			"range": map[string]interface{}{
				"@timestamp": map[string]interface{}{"gt": timestamp},
			},
		})
	}

	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": filter,
			},
		},
		"sort": []map[string]interface{}{{
			"@timestamp": map[string]interface{}{
				"order": "asc",
			},
		}},
		"_source": []string{"container.id"},
		"size":    1,
	}
}
