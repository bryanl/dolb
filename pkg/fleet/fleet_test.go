package fleet

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceName(t *testing.T) {

	examples := []struct {
		MachineID string
		Expected  string
	}{
		{MachineID: "2", Expected: "dolb-agent@1.service"},
		{MachineID: "1", Expected: "dolb-agent@2.service"},
		{MachineID: "3", Expected: ""},
	}

	for _, ex := range examples {
		name := ServiceName(ex.MachineID, validUnits)
		assert.Equal(t, ex.Expected, name)
	}
}

var (
	validUnits = fleetUnitsResponse{
		Units: []fleetUnit{
			{MachineID: "2", Name: "other-service@1.service"},
			{MachineID: "2", Name: "dolb-agent@1.service"},
			{MachineID: "1", Name: "dolb-agent@2.service"},
		},
	}
)

func TestRetrieveUnits(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/fleet/v1/units", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(&validUnits)
	}))
	defer ts.Close()

	resp, err := RetrieveUnits(ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(resp.Units))
}

func TestStopService(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		assert.Equal(t, "PUT", r.Method)

		var uscr unitStateChangeRequest
		err := json.NewDecoder(r.Body).Decode(&uscr)
		assert.NoError(t, err)
		assert.Equal(t, "inactive", uscr.DesiredState)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.WriteHeader(http.StatusNoContent)
	}))

	defer ts.Close()
	err := StopService(ts.URL, "service-1")
	assert.NoError(t, err)
}

func TestStartService(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		assert.Equal(t, "PUT", r.Method)

		var uscr unitStateChangeRequest
		err := json.NewDecoder(r.Body).Decode(&uscr)
		assert.NoError(t, err)
		assert.Equal(t, "launched", uscr.DesiredState)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.WriteHeader(http.StatusNoContent)
	}))

	defer ts.Close()
	err := StartService(ts.URL, "service-1")
	assert.NoError(t, err)
}
