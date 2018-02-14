package functest

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func TestHttpStatus(config FunctionalTestConfig) []TestFailure {
	var failures []TestFailure

	client, serverAddr, proto, err := httpClient(config)
	if err != nil {
		failures = append(failures, TestFailure{Procedure: "Status/HTTP", Message: fmt.Sprintf("HTTP client initialization error (%v)", err)})
		return failures
	}

	url := fmt.Sprintf("%s://%s/status", proto, serverAddr)

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.JsonWebToken))

	resp, err := client.Do(req)
	if err != nil {
		failures = append(failures, TestFailure{Procedure: "Status/HTTP", Message: fmt.Sprintf("HTTP GET error on %s (%v)", url, err)})
		return failures
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if fmt.Sprintf("%s", body) != "OK" {
		failures = append(failures, TestFailure{Procedure: "Stats/HTTP", Message: fmt.Sprintf("expected status \"OK\", got \"%s\"", body)})
		return failures
	}

	return failures
}
