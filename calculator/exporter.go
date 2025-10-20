package calculator

import (
	"fmt"
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

	accountID, ok := params["accountId"]
	if ok {
		err = client.SetAccountID(accountID)
		if err != nil {
			return nil, fmt.Errorf("timely.Client.SetAccountID: %w", err)
		}
	}

	return client, nil
}
