# CONTRIBUTE doc
 Fork  
 Change skillcoder to your login in github  
### GET
```
mkdir -p ~/go/src/github.com/skillcoder
cd ~/go/src/github.com/skillcoder
git clone https://github.com/skillcoder/homer.git
```

### BUILD
#### FreeBSD 11.2
```
sudo pkg install go
sudo pkg install go-glide
cd ~/go/src/github.com/skillcoder/homer
glide update
export GOOS=freebsd
gmake build
```
#### Amazon Linux 2 (AWS)
install go
```
sudo yum install glide
cd ~/go/src/github.com/skillcoder/homer
glide update
export GOOS=linux
make build
```
#### Ubuntu 18.04
```
sudo apt install golang-go
sudo apt install golang-glide
cd ~/go/src/github.com/skillcoder/homer
glide update
export GOOS=linux
make build
```
#### manual build
```
go build -race -ldflags "-X main.versionBUILD=`date -u '+%Y-%m-%d_%H:%M:%S%p'` -X main.versionCOMMIT=`git rev-parse --short HEAD` -X main.versionRELEASE=`cat VERSION`" 
```

### CONFIG
```
cp config.yml.sample config.yml
vim config.yml
```
You need install & setup clickhouse database  
create user *homer* in clickhouse-server/users.xml  
`clickhouse-client -h 127.0.0.1 -u homer --password=*secter* --query="CREATE DATABASE homer;"`  
Edit db/clickhouse.sql for your esp devices sensors  
import sql from db/clickhouse.sql to clickhouse (create tables) like this:  
```
clickhouse-client -h 127.0.0.1 -u homer --password=*secter* --database=homer < db/clickhouse.sql
```

### RUN
`./homer` or `bin/linux/homer` or `bin/freebsd/homer`  
You can change config setting by enveroment variables, see in config.yml.sample  

Example, run second instance:  
```
export export HOMER_LISTEN=127.0.0.1:18267; export HOMER_MODE=development; export HOMER_MQTT_NAME=go-homer-server-dev
./homer
```

