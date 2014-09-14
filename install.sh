yum install git hg -y

export GOROOT=/usr/local/go
export GOPATH=/data/apps/go

mkdir -p /data/apps/go/src/code.google.com/p
cd /data/apps/go/src/code.google.com/p

hg clone https://code.google.com/p/goprotobuf/
go install code.google.com/p/goprotobuf/proto
ll /usr/bin/protoc

go get git.apache.org/thrift.git/lib/go/thrift

go get github.com/citysir/zpush/join
go get github.com/citysir/zpush/node
go get github.com/citysir/zpush/offline