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
```
#### Amazon Linux 2 (AWS)
install go
```
sudo yum install glide
```
#### Ubuntu 18.04
```
sudo apt install golang-go
sudo apt install golang-glide
```
#### build
cd ~/go/src/github.com/skillcoder/homer
glide update
go build -race -ldflags "-X github.com/skillcoder/homer/version.BUILD=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X github.com/skillcoder/homer/version.COMMIT=`git rev-parse HEAD` -X github.com/skillcoder/homer/version.RELEASE=`cat VERSION`"
```

### CONFIG
```
cp config.yml.sample config.yml
vim config.yml
```

### RUN
```
export HOMER_SERVICE_LISTEN=127.0.0.1:18266; export SERVICE_MODE=development; export HOMER_SERVICE_NAME=go-homer-server
./homer
```

