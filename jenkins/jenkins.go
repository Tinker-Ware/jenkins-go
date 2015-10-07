package jenkins

import (
	"encoding/xml"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"net/http"
	"fmt"
	"bytes"
	"io/ioutil"
)

type Auth struct {
	Username string
	ApiToken string
}

type Jenkins struct {
	auth    *Auth
	baseUrl string
}

type Project struct {
	XMLName     struct{} `xml:"project"`
	Description string `xml:"description"`
}

type Job struct {
	Name  string `json:"name"`
	Url   string `json:"url"`
	Color string `json:"color"`
}
	
func NewJenkins(auth *Auth, baseUrl string) *Jenkins {
	return &Jenkins{
		auth:    auth,
		baseUrl: baseUrl,
	}
}

func (jenkins *Jenkins) buildUrl(path string, params url.Values) (requestUrl *url.URL, err error) {
	requestUrl, err = url.Parse(jenkins.baseUrl)
	if err != nil {
		return nil, err
	}

	requestUrl.Path =  path + "/api/json"
	
	if params != nil {
		queryString := params.Encode()
		if queryString != "" {
			requestUrl.RawQuery = queryString
		}
	}

	return requestUrl, nil
}

func (jenkins *Jenkins) sendRequest(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(jenkins.auth.Username, jenkins.auth.ApiToken)
	client := &http.Client{}

	return client.Do(req)
}


func (jenkins *Jenkins) get(path string, params url.Values, body interface{}) (err error) {
	requestUrl, err := jenkins.buildUrl(path, params)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("GET", requestUrl.String(), nil)
	if err != nil {
		return err
	}

	resp, err := jenkins.sendRequest(req)
	if err != nil {
		return
	}
	return jenkins.parseResponse(resp, body)
}

func (jenkins *Jenkins) parseResponse(resp *http.Response, body interface{}) (err error) {
	defer resp.Body.Close()

	if body == nil {
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	return json.Unmarshal(data, body)
}

func (jenkins *Jenkins) postXml(path string, params url.Values, xmlBody io.Reader, body interface{}) (err error) {
	requestUrl, err := url.Parse(jenkins.baseUrl)
	if err != nil {
		fmt.Println(err)
		return
	}

	requestUrl.Path = path
	
	if params != nil {
		queryString := params.Encode()
		if queryString != "" {
			requestUrl.RawQuery = queryString
		}
	}

	req, err := http.NewRequest("POST", requestUrl.String(), xmlBody)
	if err != nil {
		return
	}

	req.Header.Add("Content-Type", "application/xml")
	resp, err := jenkins.sendRequest(req)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("error: HTTP POST returned status code returned: %d", resp.StatusCode))
	}

	return jenkins.parseXmlResponse(resp, body)
}

func (jenkins *Jenkins) parseXmlResponse(resp *http.Response, body interface{}) (err error) {
	defer resp.Body.Close()

	if body == nil {
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	return xml.Unmarshal(data, body)
}


func (jenkins *Jenkins) CreateJob(project Project, jobName string) error {
	projectXml, _ := xml.Marshal(project)
	reader := bytes.NewReader(projectXml)
	params := url.Values{"name": []string{jobName}}

	return jenkins.postXml("/createItem", params, reader, nil)
}

func (jenkins *Jenkins) GetJobs() ([]Job, error) {
	var payload = struct {
		Jobs []Job `json:"jobs"`
	}{}
	err := jenkins.get("", nil, &payload)
	return payload.Jobs, err
}
