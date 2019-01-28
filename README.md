# Hooligram Server

[![Build Status](https://travis-ci.com/hooligram/hooligram-server.svg?branch=develop)](https://travis-ci.com/hooligram/hooligram-server)

## System setup

### Ubuntu 18

1. `sudo apt install wget git`
2. `wget -q https://storage.googleapis.com/golang/getgo/installer_linux`
3. `chmod +x installer_linux`
4. `./installer_linux`
5. `go version`
6. `go get github.com/hooligram/hooligram-server`
7. `cd ~/go/src/github.com/hooligram/hooligram-server`
8. `export PORT=8080`
9. `export TWILIO_API_KEY=<twilio-verify-api-key>` - [Verify](https://www.twilio.com/verify)
10. `go build && ./hooligram-server`
