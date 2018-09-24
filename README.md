# Homer
Home automation server for ESP8266 in Go.  

Features:  
 * Collecting sensors data from **all** home esp8266 nodes  
 * Every 5 sec aggregating by getting median for each  
 * Sending aggregated data in one time serie in ClickHouse database for compact and long store for future use  

### Compile and run
See in [CONTRIBUTE.md](CONTRIBUTE.md)

### HTTP Request
`curl http://127.0.0.1:18266/info`
