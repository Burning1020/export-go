/*
* Define Cumulosity sender
*/

package distro

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/edgexfoundry/export-go/pkg/models"
	"strings"
)

type mindconnectSender struct {
	client MQTT.Client
	topic  string
}

// newMindConnectSender returns new MindConnect IOT Extension sender instance.
func newMindConnectSender(addr models.Addressable) sender {
	protocol := strings.ToLower(addr.Protocol)
	broker := fmt.Sprintf("%s%s", addr.GetBaseURL(), addr.Path)
	//	deviceID := extractDeviceID(addr.Publisher)

	opts := MQTT.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(addr.Publisher)
	opts.SetUsername(addr.User)
	opts.SetPassword(addr.Password)
	opts.SetAutoReconnect(false)

	if validateProtocol(protocol) {
		c := Configuration.Certificates["MQTTS"]
		cert, err := tls.LoadX509KeyPair(c.Cert, c.Key)
		if err != nil {
			LoggingClient.Error("Failed loading x509 data")
			return nil
		}

		opts.SetTLSConfig(&tls.Config{
			ClientCAs:          nil,
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{cert},
		})
	}

	if addr.Topic == "" {
		//		addr.Topic = fmt.Sprintf("/devices/%s/events", deviceID)
		addr.Topic = "s/us"
	}

	return &mindconnectSender{
		client: MQTT.NewClient(opts),
		topic:  addr.Topic,
	}
}

func (sender *mindconnectSender) Send(data []byte, event *models.Event) bool {
	if !sender.client.IsConnected() {
		LoggingClient.Info("Connecting to mqtt server")
		token := sender.client.Connect()
		token.Wait()
		if token.Error() != nil {
			LoggingClient.Error(fmt.Sprintf("Could not connect to mqtt server, drop event. Error: %s", token.Error().Error()))
			return false
		}
	}

	//	token := sender.client.Publish(sender.topic, 0, false, data)
	var r models.Event
	json.Unmarshal(data, &r)

	createDeviceStr := "100,'IotExtDevice','IotExtDeviceType'"
	token := sender.client.Publish(sender.topic, 0, false, createDeviceStr)
	token.Wait()


	if token.Error() != nil {
		LoggingClient.Error(token.Error().Error())
		//		return false
	}

	token = sender.client.Publish(sender.topic, 0, false, "117,10")
	token.Wait()

	if token.Error() != nil {
		LoggingClient.Error(token.Error().Error())
		//		return false
	}

	for _, readingItem := range r.Readings {
		//tempPayload := [1]byte{100}
		tempPayload := fmt.Sprintf("200,pressure,P,%s,Pa", readingItem.Value)
		LoggingClient.Info(fmt.Sprintf("ReadingItem.value is %s", readingItem.Value))
		//		token := sender.client.Publish(sender.topic, 0, false, readingItem.Value)
		token := sender.client.Publish(sender.topic, 0, false, tempPayload)
		token.Wait()

		if token.Error() != nil {
			LoggingClient.Error(token.Error().Error())
			return false
		}
	}

	//
	//token.Wait()
	//
	//if token.Error() != nil {
	//	LoggingClient.Error(token.Error().Error())
	//	return false
	//}

	LoggingClient.Debug(fmt.Sprintf("Sent data: %X", data))
	return true
}

//func extractDeviceID(addr string) string {
//	return addr[strings.Index(addr, devicesPrefix)+len(devicesPrefix):]
//}


