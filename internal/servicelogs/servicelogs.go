package servicelogs

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/ergomake/ergomake/internal/database"
	"github.com/ergomake/ergomake/internal/elastic"
)

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	ServiceID string    `json:"serviceId"`
	Message   string    `json:"message"`
}

type LogStreamer interface {
	Stream(ctx context.Context, services []database.Service, namespace string, logChan chan<- []LogEntry, errChan chan<- error)
}

type esLogStreamer struct {
	elastic  elastic.ElasticSearch
	interval time.Duration
}

func NewESLogStreamer(elastic elastic.ElasticSearch, interval time.Duration) LogStreamer {
	return &esLogStreamer{elastic, interval}
}

type logsQueryResult struct {
	Hits struct {
		Hits []struct {
			Source struct {
				Timestamp time.Time `json:"@timestamp"`
				Message   string    `json:"message"`
				Log       struct {
					Offset int64 `json:"offset"`
				} `json:"log"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func (es *esLogStreamer) Stream(
	ctx context.Context,
	services []database.Service,
	namespace string,
	logChan chan<- []LogEntry,
	errChan chan<- error,
) {
	stopChans := make([]chan struct{}, len(services))
	for i := 0; i < len(services); i++ {
		stopChans[i] = make(chan struct{})
	}

	for i, service := range services {
		go func(serviceID string, i int) {
			defer func() {
				for j := 0; j < len(services); j++ {
					if i == j {
						continue
					}

					stopChans[j] <- struct{}{}
				}
			}()

			var timestamp *time.Time
			sleepDuration := time.Millisecond
			offset := int64(-1)
			currentContainer := ""
			for {
				select {
				case <-stopChans[i]:
					return
				case <-ctx.Done():
					return
				default:
				}

				nextContainer, err := es.getNextContainer(ctx, serviceID, namespace, timestamp)
				if err != nil {
					errChan <- errors.Wrap(err, "fail to get last container")
					return
				}

				if nextContainer != "" && nextContainer != currentContainer {
					currentContainer = nextContainer
					offset = -1
				}

				query := getLogsQuery(serviceID, namespace, currentContainer, offset)
				var qr logsQueryResult
				if err := es.elastic.Search(ctx, query, &qr); err != nil {
					errChan <- errors.Wrap(err, "failed to query elasticsearch")
					return
				}

				entries := make([]LogEntry, len(qr.Hits.Hits))
				for i, hit := range qr.Hits.Hits {
					entries[i] = LogEntry{
						Timestamp: hit.Source.Timestamp,
						ServiceID: serviceID,
						Message:   hit.Source.Message,
					}
					offset = hit.Source.Log.Offset
					timestamp = &hit.Source.Timestamp
				}

				select {
				case <-stopChans[i]:
					return
				case <-ctx.Done():
					return
				default:
					logChan <- entries
				}

				if len(entries) == 0 {
					sleepDuration = sleepDuration * 2
					if sleepDuration > es.interval {
						sleepDuration = es.interval
					}
				} else {
					sleepDuration = time.Millisecond * 2
				}

				time.Sleep(sleepDuration)
			}
		}(service.ID, i)
	}
}

type nextContainerQueryResult struct {
	Hits struct {
		Hits []struct {
			Source struct {
				Container struct {
					ID string `json:"id"`
				} `json:"container"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func (es *esLogStreamer) getNextContainer(
	ctx context.Context,
	serviceID string,
	namespace string,
	timestamp *time.Time,
) (string, error) {
	query := getNextContainerQuery(serviceID, namespace, timestamp)

	var qr nextContainerQueryResult
	if err := es.elastic.Search(ctx, query, &qr); err != nil {
		return "", err
	}

	if len(qr.Hits.Hits) == 0 {
		return "", nil
	}

	return qr.Hits.Hits[0].Source.Container.ID, nil
}
