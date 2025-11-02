package calculator

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/sporadisk/clocker/client/timely"
	"github.com/sporadisk/clocker/config"
	"github.com/sporadisk/clocker/event"
)

func LoadExporter(conf *config.ExporterConfig) (event.Exporter, error) {
	switch strings.ToLower(conf.Name) {
	case "timely":
		return TimelyExporter(conf.Params)
	default:
		return nil, fmt.Errorf("unrecognized exporter: %s", conf.Name)
	}
}

func getParams(params map[string]string, required ...string) (map[string]string, error) {
	result := make(map[string]string)
	for _, key := range required {
		value, ok := params[key]
		if !ok {
			return nil, fmt.Errorf("missing parameter: %s", key)
		}
		result[key] = value
	}

	return result, nil
}

func TimelyExporter(params map[string]string) (*timely.Client, error) {
	p, err := getParams(params, "applicationId", "secret", "callbackUrl")
	if err != nil {
		return nil, fmt.Errorf("getParams: %w", err)
	}

	client := &timely.Client{
		ApplicationID: p["applicationId"],
		ClientSecret:  p["secret"],
		CallbackURL:   p["callbackUrl"],
	}

	// accountId is optional
	accountID, ok := params["accountId"]
	if ok {
		accountIDInt, err := strconv.Atoi(accountID)
		if err != nil {
			return nil, fmt.Errorf("can't parse accountId as int: %w", err)
		}
		client.AccountID = accountIDInt
	}

	// projectId is required, but we handle that in the client, to make it
	// easier to find a project-id to use.
	projectId, ok := params["projectId"]
	if ok {
		projectIDInt, err := strconv.Atoi(projectId)
		if err != nil {
			return nil, fmt.Errorf("can't parse projectId as int: %w", err)
		}
		client.ProjectID = projectIDInt
	}

	err = client.Init(context.Background())
	if err != nil {
		return nil, fmt.Errorf("timely.Client.Init: %w", err)
	}

	return client, nil
}
