package parcelier

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type ErrorMessage struct {
	Code    int
	Message string
	Details interface{}
}

func NewAPI(url string) *API {
	return &API{
		url: url,
		client: &http.Client{
			Timeout: time.Duration(120 * time.Second),
		},
		verbose: true,
	}
}

type API struct {
	url         string
	queryParams *url.Values
	client      *http.Client
	verbose     bool
	veryVerbose bool
}

func (a *API) AddQueryParams(req *http.Request, queryParams map[string]string) {
	params := req.URL.Query()
	for k, v := range queryParams {
		params.Add(k, v)
	}
	req.URL.RawQuery = params.Encode()
}

func (a *API) Get(queryParams map[string]string) ([]byte, error) {
	queryURL := fmt.Sprintf("%s/query", a.url)
	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "parcelier")

	a.AddQueryParams(req, queryParams)

	if a.verbose {
		fmt.Println(req.URL.String())
	}

	///
	if a.veryVerbose {
		reqDump, err := httputil.DumpRequest(req, true)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(reqDump))
	}
	///

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	///
	if a.veryVerbose {
		respDump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(respDump))
	}
	///

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	status := resp.StatusCode
	switch status {
	case 400:
		return nil, fmt.Errorf("error: 400 : %s", string(body))
	case 404:
		return nil, fmt.Errorf("error: 404 : %s", string(body))
	}
	return body, nil
}