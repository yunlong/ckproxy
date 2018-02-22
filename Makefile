CURDIR=$(shell pwd)
SVNVERSION  = 1.1.$(shell expr `git rev-list --all|wc -l` + 0)

export GOARCH=amd64
export GOOS=darwin
#export GOOS=linux
#export GOOS=windows

all:
	go get github.com/golang/glog
	go get github.com/zieckey/goini
	GO15VENDOREXPERIMENT=1	GOPATH=$(CURDIR):$(GOPATH) go build -x  -o $(CURDIR)/sbin/ckproxy.${SVNVERSION} $(CURDIR)/src/main/main.go
	GO15VENDOREXPERIMENT=1	GOPATH=$(CURDIR):$(GOPATH) go build -x  -o $(CURDIR)/sbin/demo.${SVNVERSION} $(CURDIR)/src/demo/main.go
	ln -sf $(CURDIR)/sbin/ckproxy.${SVNVERSION} $(CURDIR)/sbin/ckproxy

clean:
	rm -rf sbin/ckproxy
	rm -rf sbin/ckproxy.1.*
	rm -rf sbin/demo*
	rm -rf skylarproxy
	rm -rf skylarproxy*.tar.gz

t:
	echo test

pub:
	echo pub

test:
	cd ${CURDIR}/src/ckproxy/;go test

pkg:all
	mkdir -p skylarproxy skylarproxy/conf  skylarproxy/sbin/ skylarproxy/logs/
	cp $(CURDIR)/conf/app.conf	skylarproxy/conf/
	cp $(CURDIR)/sbin/srvctl	skylarproxy/sbin/
	cp $(CURDIR)/sbin/*.cron	skylarproxy/sbin/
	cp $(CURDIR)/sbin/*.sh		skylarproxy/sbin/
	cp $(CURDIR)/sbin/ckproxy.${SVNVERSION} skylarproxy/sbin/ckproxy
	#ln -sf skylarproxy/sbin/ckproxy.${SVNVERSION} skylarproxy/sbin/ckproxy
	echo ${SVNVERSION} > skylarproxy/VERSION
	tar -zcvf skylarproxy.${SVNVERSION}.tar.gz  skylarproxy
