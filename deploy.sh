#!/bin/bash

deploy() {
  echo "deploying..."
  if [[ ! -f ./hooligram-developer.pem ]]; then
    echo "error: hooligram-developer.pem file is missing"
    return 1
  fi

  ip_addr="$(echo $IP_ADDR)"

  if [[ -z $ip_addr ]]; then
    echo "error: IP_ADDR env variable is missing"
    return 1
  fi

  echo "server ip address: $ip_addr"
  ssh -i ./hooligram-developer.pem "ubuntu@$ip_addr" "cd ~/go/src/github.com/hooligram/hooligram-server && git checkout master && git pull && /home/ubuntu/.go/bin/go get -d ./... && /home/ubuntu/.go/bin/go install && sudo service hooligram restart && service hooligram status"

  return 0
}

if deploy; then
  echo "deployment done!"
else
  echo "deployment failed :("
fi
