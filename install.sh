yum install git hg -y

mkdir -p /data/apps/gopath/src/code.google.com/p
cd /data/apps/gopath/src/code.google.com/p

hg clone https://code.google.com/p/goprotobuf/
go install code.google.com/p/goprotobuf/proto
ll /usr/bin/protoc

go get git.apache.org/thrift.git/lib/go/thrift
go get github.com/citysir/zpush