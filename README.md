# Hooligram Server

**Ubuntu 18 server setup**

1. `sudo apt install wget git`
2. `wget -q https://storage.googleapis.com/golang/getgo/installer_linux`
3. `chmod +x installer_linux`
4. `./installer_linux`
5. `go version`
6. `go get github.com/hooligram/hooligram-server`
7. `cd ~/go/src/github.com/hooligram/hooligram-server`
8. `go build && ./hooligram-server`
