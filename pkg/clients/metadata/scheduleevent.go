/*******************************************************************************
 * Copyright 2018 Dell Inc.
 * Copyright (C) 2018 Canonical Ltd
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/
package metadata

import (
	"encoding/json"
	"net/url"

	"github.com/edgexfoundry/export-go/pkg/clients"
	"github.com/edgexfoundry/export-go/pkg/clients/types"
	"github.com/edgexfoundry/export-go/pkg/models"
)

// ScheduleEventClient is an interface used to operate on Core Metadata schedule event objects.
type ScheduleEventClient interface {
	Add(dev *models.ScheduleEvent) (string, error)
	Delete(id string) error
	DeleteByName(name string) error
	ScheduleEvent(id string) (models.ScheduleEvent, error)
	ScheduleEventForName(name string) (models.ScheduleEvent, error)
	ScheduleEvents() ([]models.ScheduleEvent, error)
	ScheduleEventsForAddressable(name string) ([]models.ScheduleEvent, error)
	ScheduleEventsForAddressableByName(name string) ([]models.ScheduleEvent, error)
	ScheduleEventsForServiceByName(name string) ([]models.ScheduleEvent, error)
	Update(dev models.ScheduleEvent) error
}

// ScheduleEventRestClient is struct used as a receiver for ScheduleEventClient interface methods.
type ScheduleEventRestClient struct {
	url      string
	endpoint clients.Endpointer
}

// NewScheduleEventClient returns a new instance of ScheduleEventClient.
func NewScheduleEventClient(params types.EndpointParams, m clients.Endpointer) ScheduleEventClient {
	s := ScheduleEventRestClient{endpoint: m}
	s.init(params)
	return &s
}

func (s *ScheduleEventRestClient) init(params types.EndpointParams) {
	if params.UseRegistry {
		ch := make(chan string, 1)
		go s.endpoint.Monitor(params, ch)
		go func(ch chan string) {
			for {
				select {
				case url := <-ch:
					s.url = url
				}
			}
		}(ch)
	} else {
		s.url = params.Url
	}
}

// Helper method to request and decode a schedule event
func (s *ScheduleEventRestClient) requestScheduleEvent(url string) (models.ScheduleEvent, error) {
	data, err := clients.GetRequest(url)
	if err != nil {
		return models.ScheduleEvent{}, err
	}

	se := models.ScheduleEvent{}
	err = json.Unmarshal(data, &se)
	return se, err
}

// Helper method to request and decode a schedule event slice
func (s *ScheduleEventRestClient) requestScheduleEventSlice(url string) ([]models.ScheduleEvent, error) {
	data, err := clients.GetRequest(url)
	if err != nil {
		return []models.ScheduleEvent{}, err
	}

	seSlice := make([]models.ScheduleEvent, 0)
	err = json.Unmarshal(data, &seSlice)
	return seSlice, err
}

// Add a schedule event.
func (s *ScheduleEventRestClient) Add(se *models.ScheduleEvent) (string, error) {
	return clients.PostJsonRequest(s.url, se)
}

// Delete a schedule event (specified by id).
func (s *ScheduleEventRestClient) Delete(id string) error {
	return clients.DeleteRequest(s.url + "/id/" + id)
}

// Delete a schedule event (specified by name).
func (s *ScheduleEventRestClient) DeleteByName(name string) error {
	return clients.DeleteRequest(s.url + "/name/" + url.QueryEscape(name))
}

// ScheduleEvent returns the ScheduleEvent specified by id.
func (s *ScheduleEventRestClient) ScheduleEvent(id string) (models.ScheduleEvent, error) {
	return s.requestScheduleEvent(s.url + "/" + id)
}

// ScheduleEventForName returns the ScheduleEvent specified by name.
func (s *ScheduleEventRestClient) ScheduleEventForName(name string) (models.ScheduleEvent, error) {
	return s.requestScheduleEvent(s.url + "/name/" + url.QueryEscape(name))
}

// Get a list of all schedules events.
func (s *ScheduleEventRestClient) ScheduleEvents() ([]models.ScheduleEvent, error) {
	return s.requestScheduleEventSlice(s.url)
}

// ScheduleEventForAddressable returns the ScheduleEvent specified by addressable.
func (s *ScheduleEventRestClient) ScheduleEventsForAddressable(addressable string) ([]models.ScheduleEvent, error) {
	return s.requestScheduleEventSlice(s.url + "/addressable/" + url.QueryEscape(addressable))
}

// ScheduleEventForAddressableByName returns the ScheduleEvent specified by addressable name.
func (s *ScheduleEventRestClient) ScheduleEventsForAddressableByName(name string) ([]models.ScheduleEvent, error) {
	return s.requestScheduleEventSlice(s.url + "/addressablename/" + url.QueryEscape(name))
}

// Get the schedule event for service by name.
func (s *ScheduleEventRestClient) ScheduleEventsForServiceByName(name string) ([]models.ScheduleEvent, error) {
	return s.requestScheduleEventSlice(s.url + "/servicename/" + url.QueryEscape(name))
}

// Update a schedule event - handle error codes
func (s *ScheduleEventRestClient) Update(se models.ScheduleEvent) error {
	return clients.UpdateRequest(s.url, se)
}
