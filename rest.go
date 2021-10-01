package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/imroc/req"
)

type InputDetail struct {
	Name       string `json:"name"`
	Identifier string `json:"id"`
	State      string `json:"state"`
	Message    string `json:"message"`
}

type InputOverview []struct {
	ID string `json:"id"`
}

type CloudService struct {
	Name  string `json:"service"`
	State string `json:"state"`
}

type MQTTRestResponse struct {
	Default    string `json:"defaultBroker"`
	Kubernetes string `json:"kubernetes"`
}

type MQTTServer struct {
	Name  string
	State string
}

// HasStatus checks if the alarm input has the given state
func (i InputDetail) HasStatus(status string) float64 {
	if i.State == status {
		return 1
	} else {
		return 0
	}
}

// HasStatus checks if the cloudservice has the given state
func (c CloudService) HasStatus(status string) float64 {
	if c.State == status {
		return 1
	} else {
		return 0
	}
}

// HasStatus checks if the cloudservice has the given state
func (m MQTTServer) HasStatus(status string) float64 {
	if m.State == status {
		return 1
	} else {
		return 0
	}
}

// GetValue searches in the message field for a possible numerical value
func (i InputDetail) GetValue() (float64, error) {
	r, _ := regexp.Compile("[0-9]*,[0-9[0-9]]*")
	if v := r.FindString(i.Message); v != "" {

		// Convert float from german to american notation
		v = strings.ReplaceAll(v, ",", ".")

		f, _ := strconv.ParseFloat(v, 64)
		return f, nil
	} else {
		return 0, errors.New("no value in message")
	}
}

// QueryInputs returns a list with informations about all alarm inputs
func QueryInputs(hostname string, accessKey string) (*[]InputDetail, error) {
	authHeader := req.Header{
		"Accept":        "application/json",
		"Authorization": accessKey,
	}
	var inputDetails []InputDetail
	var inputOverview InputOverview

	r, err := req.Get(fmt.Sprintf("%s/rest/monitoring/input", hostname), authHeader)
	if err != nil {
		return nil, err
	}

	err = r.ToJSON(&inputOverview)
	if err != nil {
		return nil, err
	}

	for _, input := range inputOverview {
		r, err := req.Get(fmt.Sprintf("%s/rest/monitoring/input/%s", hostname, input.ID), authHeader)
		if err != nil {
			return nil, err
		}

		var detail InputDetail
		err = r.ToJSON(&detail)
		if err != nil {
			return nil, err
		}

		detail.Identifier = input.ID
		inputDetails = append(inputDetails, detail)
	}

	return &inputDetails, nil
}

// QueryCloudServices returns a list of all cloudservices
func QueryCloudServices(hostname string, accessKey string) (*[]CloudService, error) {
	authHeader := req.Header{
		"Accept":        "application/json",
		"Authorization": accessKey,
	}
	var services []CloudService

	r, err := req.Get(fmt.Sprintf("%s/rest/monitoring/cloud", hostname), authHeader)
	if err != nil {
		return nil, err
	}

	err = r.ToJSON(&services)
	if err != nil {
		return nil, err
	}

	return &services, nil
}

// QueryMQTTServer returns a list of all mqtt brokers
func QueryMQTTServer(hostname string, accessKey string) (*[]MQTTServer, error) {
	authHeader := req.Header{
		"Accept":        "application/json",
		"Authorization": accessKey,
	}
	var resp MQTTRestResponse

	r, err := req.Get(fmt.Sprintf("%s/rest/monitoring/mqtt", hostname), authHeader)
	if err != nil {
		return nil, err
	}

	err = r.ToJSON(&resp)
	if err != nil {
		return nil, err
	}

	// Convert data structure from single object to list of objects
	// to fit the structure of all other endpoints
	mqttServer := []MQTTServer{
		MQTTServer{
			Name:  "defaultBroker",
			State: resp.Default,
		},
		MQTTServer{
			Name:  "kubernetes",
			State: resp.Kubernetes,
		},
	}

	return &mqttServer, nil
}
