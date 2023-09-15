package metricdescriptionprocessor

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

func TestProcessMetrics(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{}
	processor := newMetricsDescriptionProcessor(logger, config)

	md := pmetric.NewMetrics()

	_, err := processor.processMetrics(context.Background(), md)

	assert.NoError(t, err)
	// Add more assertions based on your expected results
}

func TestUpdateColumnKeyMap(t *testing.T) {
	// Create a logger for testing
	logger, _ := zap.NewDevelopment()

	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		assert.Equal(t, req.URL.String(), "/1/columns/testDataset?key_name=testKey")
		// Send response to be tested
		rw.Write([]byte(`{"id": "testId"}`))
	}))
	// Close the server when test finishes
	defer server.Close()

	// Create a new metricDescriptionProcessor
	m := &metricDescriptionProcessor{
		logger:      logger,
		lookupTable: make(map[string]string),
		columnCache: make(map[string]string),
		cfg: &Config{
			Dataset: "testDataset",
			APIKey:  "testAPIKey",
		},
	}

	// Update the endpoint in the function to point to the mock server
	endpoint := fmt.Sprintf("%s/1/columns/%s?key_name=%s", server.URL, m.cfg.Dataset, "testKey")

	// Call the function to test
	m.updateColumnKeyMap("testKey", endpoint)

	// Assert that the columnCache was updated correctly
	assert.Equal(t, "testId", m.columnCache["testKey"])
}

func TestUpdateDescriptionByKey(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check method and path
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/1/columns/testDataset/testCid", r.URL.Path)

		// Check headers
		assert.Equal(t, "testApiKey", r.Header.Get("X-Honeycomb-Team"))

		// Check body
		body, _ := ioutil.ReadAll(r.Body)
		assert.Equal(t, `{"description":"testDescription"}`, strings.Trim(string(body), "\n"))

		// Respond with OK
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Create a processor with the test server's URL as the endpoint
	p := &metricDescriptionProcessor{
		logger:      zap.NewNop(),
		cfg:         &Config{APIKey: "testApiKey", Dataset: "testDataset"},
		lookupTable: make(map[string]string),
		columnCache: map[string]string{"testKey": "testCid"},
	}

	// Call the function
	p.updateDescriptionByKey("testKey", "testDescription", ts.URL+"/1/columns/testDataset/testCid")
}
