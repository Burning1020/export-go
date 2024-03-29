/*******************************************************************************
 * Copyright 2017 Dell Inc.
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
package messaging

import (
	"github.com/edgexfoundry/export-go/pkg/models"
)

// Configuration struct for PubSub
type PubSubConfiguration struct {
	AddressPort string
}

type EventPublisher interface {
	SendEventMessage(e models.Event) error
}

func NewEventPublisher(conf PubSubConfiguration) EventPublisher {
	return newZeroMQEventPublisher(conf)
}
