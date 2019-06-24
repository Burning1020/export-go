/*
* Define Open Edge Device Kit sender
* Do preparation work(connect to mosquitto, init oedk, store DataSorceId and DataPointId)
* Send data
*/

package distro

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/edgexfoundry/export-go/pkg/models"
	"strings"
	"time"
)

type oedkconnectSender struct {
	client		MQTT.Client
	config		*configuration
	GetId		bool
	Onboard		bool
}

// newMindConnectSender returns new MindConnect IOT Extension sender instance.
func newOEDKConnectSender(addr models.Addressable) *oedkconnectSender {
	protocol := strings.ToLower(addr.Protocol)
	broker := fmt.Sprintf("%s%s", addr.GetBaseURL(), addr.Path)

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
	conf, err := LoadConfigFromFile()
	if err != nil {
		panic(fmt.Errorf("Read Open Edge Device Kit configuration failed: %v", err))
	}
	return &oedkconnectSender{
		client:		MQTT.NewClient(opts),
		config:		conf,
		GetId:		false,
		Onboard:	false,
	}
}

// do some preparations
func (sender *oedkconnectSender) Prepare() {
	// connect to mosquitto server(local mqtt broker)
	LoggingClient.Info("Connecting to mosquitto")
	token := sender.client.Connect()
	if token.Wait() && token.Error() != nil {
		LoggingClient.Error(fmt.Sprintf("Could not connect to mosquitto, drop event. Error: %s", token.Error().Error()))
		panic(token.Error())
	}
	LoggingClient.Info("Connected into mosquitto")

	// initialize agent iff client is connected for the first time
	if !sender.config.Oedk.IsInitialized {
		token = sender.client.Publish(INIT_TOPIC, 0, false, sender.config.Oedk.InitJson)
		if token.Wait() && token.Error() != nil {
			LoggingClient.Error(token.Error().Error())
		}

		go func() {
			// init failed?
			token = sender.client.Subscribe(INITINFO_TOPIC, 0, sender.HandleInitInfo)
			if token.Wait() && token.Error() != nil {
				LoggingClient.Error(token.Error().Error())
			}
			select {}
		}()
	}

	// subscribe agentruntime/monitoring/diagnostic/onboarding
	//           agentruntime/monitoring/diagnostic/connection
	go func() {

		token = sender.client.Subscribe(ONBOARDING_TOPIC, 0, sender.HandleOnBoardTopic)
		if token.Wait() && token.Error() != nil {
			LoggingClient.Error(token.Error().Error())
		}

		select {}
	}()


	// subscribe cloud/monitoring/update/configuration/{protocol}
	go func() {
		ProTopic := strings.Replace(CONFIGPROINFO_TOPIC, "{protocol}", "OPCUA", 1)
		token = sender.client.Subscribe(ProTopic, 0, sender.HandleProTopic)
		if token.Wait() && token.Error() != nil {
			LoggingClient.Error(token.Error().Error())
		}

		select {}
	}()

	if sender.config.DataSource.Id == "" {
		sender.GetId = false
		return
	}
	for _, dp := range sender.config.DataSource.DataPoint {
		if dp.Id != "" {
			sender.GetId = true
			LoggingClient.Info("[Open Edge Device Kit] Get GetId for uploading date")
			return
		}
	}
}

func (sender *oedkconnectSender) Send(data []byte, event *models.Event) bool {
	if !sender.client.IsConnected() {
		LoggingClient.Info("Connecting to mosquitto")
		token := sender.client.Connect()
		token.Wait()
		if token.Error() != nil {
			LoggingClient.Error(fmt.Sprintf("Could not connect to mosquitto, drop event. Error: %s", token.Error().Error()))
			return false
		}
		LoggingClient.Info("Connected into mosquitto")
	}

	if sender.GetId && sender.Onboard {
		var r models.Event
		json.Unmarshal(data, &r)
		var index = int(0) // the number of DataPointId that linked with data

		for _, readingItem := range r.Readings {
			LoggingClient.Info(fmt.Sprintf("ReadingItem.value is %s", readingItem.Value))
			//token := sender.client.Publish(sender.topic, 0, false, tempPayload)

			jsonTemplate := "[\n" +
				"  {\n" +
				"    \"timestamp\": \"{timeStamp}\",\n" +
				"    \"values\": [\n" +
				"      {\n" +
				"        \"dataPointId\": \"{dataPointId}\",\n" +
				"        \"value\": \"{dataValue}\",\n" +
				"        \"qualityCode\": \"{qualityCode}\"\n" +
				"      }\n" +
				"    ]\n" +
				"  }\n" +
				"]";
			jsonTemplate = strings.Replace(jsonTemplate, "{timeStamp}", time.Now().Format(time.RFC3339), 1)
			jsonTemplate = strings.Replace(jsonTemplate, "{dataPointId}", sender.config.DataSource.DataPoint[index].Id, 1)
			jsonTemplate = strings.Replace(jsonTemplate, "{dataValue}", readingItem.Value, 1)
			jsonTemplate = strings.Replace(jsonTemplate, "{qualityCode}", "0", 1)

			DataTopic := strings.Replace(UPLOADDATA_TOPIC, "{protocol}", "OPCUA", 1)
			DataTopic = strings.Replace(DataTopic, "{dataSourceId}", sender.config.DataSource.Id, 1)
			token := sender.client.Publish(DataTopic, 0, false, jsonTemplate)

			if token.Wait() && token.Error() != nil {
				LoggingClient.Error(token.Error().Error())
				return false
			}
		}
		LoggingClient.Debug(fmt.Sprintf("Sent data"))
		return true
	}
	return false
}
// transform Mqtt message to map struct
func (sender *oedkconnectSender) HandleInitInfo(client MQTT.Client, message MQTT.Message) {
	var response map[string]interface{}
	json.Unmarshal(message.Payload(), &response)

	value := response["value"].(float64)
	status := response["status"].(string)

	if value == 1 {
		LoggingClient.Error(fmt.Sprintf("[Open Edge Device Kit] %s", status))
		sender.config.Oedk.IsInitialized = false
		return
	}
	sender.config.Oedk.IsInitialized = true
	LoggingClient.Debug("[Open Edge Device Kit] Init Successful")

	// update config file
	UpdateConfigFromFile(sender.config)
}
// transform Mqtt message to map struct
func (sender *oedkconnectSender) HandleOnBoardTopic(client MQTT.Client, message MQTT.Message) {
	var response map[string]interface{}
	json.Unmarshal(message.Payload(), &response)

	value := response["value"].(float64)
	state := response["state"].(string)

	if value == 2{  // in progress
		LoggingClient.Info(fmt.Sprintf("[Open Edge Device Kit] %s", state))
		return
	}
	if value == 3 || value == 4 { // failed or offboarded
		LoggingClient.Error(fmt.Sprintf("[Open Edge Device Kit] %s", state))
		return
	}
	if value == 1 { // success
		LoggingClient.Info(fmt.Sprintf("[Open Edge Device Kit] %s", state))
		sender.Onboard = true
		return
	}

}

// transform Mqtt message to map struct
func (sender *oedkconnectSender) HandleProTopic(client MQTT.Client, message MQTT.Message) {
	LoggingClient.Info("[Open Edge Device Kit] Get latest configuration from Mindsphere")
	var response map[string]interface{}
	json.Unmarshal(message.Payload(), &response)

	sender.config.DataSource.Id = response["dataSourceId"].(string) // DataSourceId

	DataPoint, ok := response["dataPoints"];
	if  !ok {
		LoggingClient.Error("[Open Edge Device Kit] Please add at least one DataPoint in your DataSource")
		return
	}

	// DatePoints array
	for i, DataPointItem := range DataPoint.([]interface{}) {
		DataPoint := DataPointItem.(map[string]interface{})
		sender.config.DataSource.DataPoint[i].Id = DataPoint["dataPointId"].(string)
	}
	sender.GetId = true
	LoggingClient.Info("[Open Edge Device Kit] Get GetId for uploading date")

	// update config file
	UpdateConfigFromFile(sender.config)
}
