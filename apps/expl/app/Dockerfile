from expl-base
run mkdir -p /root/.ssh/
run mkdir -p /root/go/src/github.com/fmd/
run echo "Host github.com\n\tStrictHostKeyChecking no\n" >> /root/.ssh/config
run git clone https://github.com/fmd/micro-app /root/go/src/github.com/fmd/micro-app
run ln -s /root/go/src/github.com/fmd/micro-app /app
workdir /app

env GOPATH /root/go

run go get github.com/onsi/ginkgo
run go get github.com/onsi/gomega
run go get .
run go build .
run go test .
