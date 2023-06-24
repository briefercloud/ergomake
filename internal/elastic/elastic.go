package elastic

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/pkg/errors"
)

type ElasticSearch interface {
	Search(ctx context.Context, query any, result any) error
}

type client struct {
	*elasticsearch.Client
}

func NewElasticSearch(addr string, username string, password string) (ElasticSearch, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{addr},
		Username:  username,
		Password:  password,
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &client{es}, nil
}

func (client *client) Search(ctx context.Context, query any, result any) error {
	var queryBuf bytes.Buffer
	if err := json.NewEncoder(&queryBuf).Encode(query); err != nil {
		return errors.Wrap(err, "failed to json encode elasticsearch query")
	}

	res, err := client.Client.Search(
		client.Client.Search.WithContext(ctx),
		client.Client.Search.WithIndex(".ds-filebeat*"),
		client.Client.Search.WithBody(&queryBuf),
		client.Client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return errors.Wrap(err, "failed to query elasticsearch")
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return errors.Wrap(err, "failed to parse elasticsearch error response body")
		}

		err := errors.Errorf("[%s] %s: %s",
			res.Status(),
			e["error"].(map[string]interface{})["type"],
			e["error"].(map[string]interface{})["reason"],
		)

		return errors.Wrap(err, "failed to query elasticsearch")
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return errors.Wrap(err, "failed to parse elasticsearch response body")
	}

	return nil
}
