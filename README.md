# Homer
Home automation server for ESP8266 in Go.

### Run
Run server:
```
go build; export MQTT_SERVER=192.168.100.254:1883; export HOMER_SERVICE_LISTEN=192.168.100.254:18266; export SERVICE_MODE=development; export HOMER_SERVICE_NAME=go-homer-server
./homer
```

### HTTP Request
`curl http://192.168.100.254:18266/info`
