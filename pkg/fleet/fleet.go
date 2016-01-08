package fleet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
)

var reDolbService = regexp.MustCompile(`^dolb-agent@\d+\.service$`)

type fleetUnitsResponse struct {
	Units []fleetUnit `json:"units"`
}

type fleetUnit struct {
	MachineID string `json:"machineID"`
	Name      string `json:"name"`
}

type unitStateChangeRequest struct {
	DesiredState string `json:"desiredState"`
}

func ServiceName(machineID string, unitsResp fleetUnitsResponse) string {
	for _, unit := range unitsResp.Units {
		if unit.MachineID == machineID && reDolbService.Match([]byte(unit.Name)) {
			return unit.Name
		}
	}

	return ""
}

func RetrieveUnits(baseURL string) (fleetUnitsResponse, error) {
	var fur fleetUnitsResponse
	u, err := url.Parse(baseURL)
	if err != nil {
		return fur, err
	}

	u.Path = "/fleet/v1/units"
	resp, err := http.Get(u.String())
	if err != nil {
		return fur, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fur, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&fur)
	if err != nil {
		return fur, err
	}

	return fur, nil
}

func StopService(baseURL, service string) error {
	return changeServiceState(baseURL, service, "inactive")
}

func StartService(baseURL, service string) error {
	return changeServiceState(baseURL, service, "launched")
}

func changeServiceState(baseURL, service, state string) error {
	uscr := unitStateChangeRequest{DesiredState: state}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(&uscr)

	u, err := url.Parse(baseURL)
	if err != nil {
		return err
	}

	u.Path = fmt.Sprintf("/fleet/v1/units/%s", service)
	req, err := http.NewRequest("PUT", u.String(), &buf)
	if err != nil {
		return err
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 204 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
