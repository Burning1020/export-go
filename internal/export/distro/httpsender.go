//
// Copyright (c) 2017
// Cavium
// Mainflux
// IOTech
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"

	"github.com/edgexfoundry/export-go/pkg/models"
)

type httpSender struct {
	url    string
	method string
}

const mimeTypeJSON = "application/json"

// newHTTPSender - create http sender
func newHTTPSender(addr models.Addressable) sender {

	sender := httpSender{
		url:    addr.Protocol + "://" + addr.Address + ":" + strconv.Itoa(addr.Port) + addr.Path,
		method: addr.HTTPMethod,
	}
	return sender
}

func (sender httpSender) Send(data []byte, event *models.Event) bool {

	switch sender.method {
	case http.MethodPost:
		response, err := http.Post(sender.url, mimeTypeJSON, bytes.NewReader(data))
		if err != nil {
			LoggingClient.Error(err.Error())
			return false
		}
		defer response.Body.Close()
		LoggingClient.Info(fmt.Sprintf("Response: %s", response.Status))
	default:
		LoggingClient.Info(fmt.Sprintf("Unsupported method: %s", sender.method))
		return false
	}

	LoggingClient.Info(fmt.Sprintf("Sent data: %X", data))
	return true
}
