export GOPATH=/data/apps/go

yum install git hg -y

mkdir -p /data/apps/go/src/code.google.com/p
cd /data/apps/go/src/code.google.com/p

hg clone https://code.google.com/p/goprotobuf
hg clone https://code.google.com/p/log4go

cd /data/apps/go/src/code.google.com/p/goprotobuf
make

go get git.apache.org/thrift.git/lib/go/thrift

go get github.com/citysir/zpush/join
go get github.com/citysir/zpush/node
go get github.com/citysir/zpush/offline