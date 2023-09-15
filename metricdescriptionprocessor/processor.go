package metricdescriptionprocessor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"go.uber.org/zap"
)

// This struct contains a map that will act as your in-memory lookup table.
type metricDescriptionProcessor struct {
	logger        *zap.Logger
	cfg           *Config
	lookupTable   map[string]string
	columnCache   map[string]string
	refreshNeeded bool
}

func newMetricsDescriptionProcessor(logger *zap.Logger, config component.Config) *metricDescriptionProcessor {
	m := &metricDescriptionProcessor{
		logger:        logger,
		lookupTable:   make(map[string]string),
		columnCache:   make(map[string]string),
		cfg:           config.(*Config),
		refreshNeeded: true,
	}
	return m
}

func (m *metricDescriptionProcessor) processMetrics(ctx context.Context, md pmetric.Metrics) (pmetric.Metrics, error) {
	rms := md.ResourceMetrics()
	for i := 0; i < rms.Len(); i++ {
		rs := rms.At(i)
		ilms := rs.ScopeMetrics()
		for j := 0; j < ilms.Len(); j++ {
			ils := ilms.At(j)
			metrics := ils.Metrics()
			for k := 0; k < metrics.Len(); k++ {
				metric := metrics.At(k)
				if metric.Description() != "" {
					if desc, ok := m.lookupTable[metric.Name()]; !ok || desc != metric.Description() {
						m.lookupTable[metric.Name()] = metric.Description()
						m.refreshNeeded = true
					}
				}
			}
		}
	}
	return md, nil
}

func (m *metricDescriptionProcessor) startUpdateLoop() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			m.doUpdate()
		}
	}()
}

func (m *metricDescriptionProcessor) doUpdate() {
	if m.refreshNeeded {
		for key, description := range m.lookupTable {
			m.updateColumnKeyMap(key, "")
			m.updateDescriptionByKey(key, description, "")
		}
		m.refreshNeeded = false
	}
}

func (m *metricDescriptionProcessor) updateColumnKeyMap(key string, endpoint string) {
	if _, exists := m.columnCache[key]; !exists {
		if endpoint == "" {
			endpoint = fmt.Sprintf("https://api.honeycomb.io/1/columns/%s?key_name=%s", m.cfg.Dataset, key)
		}
		client := &http.Client{}
		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			m.logger.Error("Error creating new request", zap.Error(err))
			return
		}
		req.Header.Add("X-Honeycomb-Team", m.cfg.APIKey)
		resp, err := client.Do(req)
		if err != nil {
			m.logger.Error("Error making request", zap.Error(err))
			return
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			m.logger.Error("Error reading response body", zap.Error(err))
			return
		}

		if resp.StatusCode == http.StatusOK {
			var result map[string]interface{}
			json.Unmarshal(body, &result)
			if id, ok := result["id"]; ok {
				m.columnCache[key] = id.(string)
			}
		} else if resp.StatusCode == http.StatusNotFound {
			m.logger.Info("Key did not exist in dataset")
		} else {
			m.logger.Info("Unsuccessful", zap.Int("Status Code", resp.StatusCode))
		}

	}
}

func (m *metricDescriptionProcessor) updateDescriptionByKey(key string, description string, endpoint string) {
	cid := m.columnCache[key]
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://api.honeycomb.io/1/columns/%s/%s", m.cfg.Dataset, cid)
		m.logger.Info(endpoint)
	}
	client := &http.Client{}
	req, err := http.NewRequest("PUT", endpoint, nil)
	if err != nil {
		m.logger.Error("Error creating new request", zap.Error(err))
		return
	}
	req.Header.Add("X-Honeycomb-Team", m.cfg.APIKey)
	jsonData := map[string]string{
		"description": description,
	}
	jsonValue, _ := json.Marshal(jsonData)
	req.Body = ioutil.NopCloser(bytes.NewBuffer(jsonValue))
	resp, err := client.Do(req)
	if err != nil {
		m.logger.Error("Error making request", zap.Error(err))
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		m.logger.Error("Non-OK HTTP status:", zap.String("status", resp.Status))
		return
	}
}
