# Hooligram Server

[![Build Status](https://travis-ci.com/hooligram/hooligram-server.svg?branch=develop)](https://travis-ci.com/hooligram/hooligram-server)

## System setup (Ubuntu 18)

### Go

1. `sudo apt install wget git`
2. `wget -q https://storage.googleapis.com/golang/getgo/installer_linux`
3. `chmod +x installer_linux`
4. `./installer_linux`
5. `go version`
6. `go get github.com/hooligram/hooligram-server`
7. `cd ~/go/src/github.com/hooligram/hooligram-server`
8. `export PORT=8080`
9. `export TWILIO_API_KEY=<twilio-verify-api-key>` - [Verify](https://www.twilio.com/verify)
10. `export MYSQL_DB_NAME=hooligram`
11. `export MYSQL_USERNAME=<username>`
12. `export MYSQL_PASSWORD=<password>`
13. `go build && ./hooligram-server`

### MySQL DB

1. `sudo apt update`
2. `sudo apt install mysql-server`
3. `sudo systemctl status mysql` - Make sure the *Active* status is *active (running)*
4. `sudo mysql #opens mysql console as root`
5. `CREATE USER '<username>'@'localhost' IDENTIFIED BY '<password>';`
6. `GRANT ALL PRIVILEGES ON hooligram.* TO '<username>'@'localhost' IDENTIFIED BY '<password>';`
7. `CREATE DATABASE hooligram;`

