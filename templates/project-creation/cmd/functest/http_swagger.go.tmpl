package functest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func TestHttpSwagger(config FunctionalTestConfig) []TestFailure {
	var failures []TestFailure

	client, serverAddr, proto, err := httpClient(config)
	if err != nil {
		failures = append(failures, TestFailure{Procedure: "Swagger/HTTP", Message: fmt.Sprintf("HTTP client initialization error (%v)", err)})
		return failures
	}

	url := fmt.Sprintf("%s://%s/api/v1/swagger.json", proto, serverAddr)

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.JsonWebToken))
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		failures = append(failures, TestFailure{Procedure: "Swagger/HTTP", Message: fmt.Sprintf("HTTP GET error on %s (%v)", url, err)})
		return failures
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var responseData interface{}
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		failures = append(failures, TestFailure{Procedure: "Swagger/HTTP", Message: fmt.Sprintf("JSON parsing error (%v)", err)})
		return failures
	}

	responseMap, ok := responseData.(map[string]interface{})
	if !ok {
		failures = append(failures, TestFailure{Procedure: "Swagger/HTTP", Message: "JSON parsing error (top-level map)"})
		return failures
	}

	_, ok = responseMap["swagger"].(string)
	if !ok {
		failures = append(failures, TestFailure{Procedure: "Swagger/HTTP", Message: "JSON response should contain the Swagger version field"})
		return failures
	}

	_, ok = responseMap["info"].(map[string]interface{})
	if !ok {
		failures = append(failures, TestFailure{Procedure: "Swagger/HTTP", Message: "JSON response should contain the Swagger info object"})
		return failures
	}

	_, ok = responseMap["paths"].(map[string]interface{})
	if !ok {
		failures = append(failures, TestFailure{Procedure: "Swagger/HTTP", Message: "JSON response should contain the Swagger paths object"})
		return failures
	}

	return failures
}
