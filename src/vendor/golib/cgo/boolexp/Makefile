# set CPUS for Linux or FreeBSD
	$(PROTO_BASE)$(PROTO) --python_out=$(OUTPUT_DIR) --cpp_out=$(OUTPUT_DIR)  --go_out=$(OUTPUT_DIR) $< 
PLATFORM := $(shell uname)
CPUS := $(strip $(if $(shell echo $(PLATFORM)|grep Linux),\
	$(shell cat /proc/cpuinfo|grep -c processor),\
	$(shell sysctl -a | egrep -i 'hw.ncpu' | cut -d: -f2)))


CC := gcc
CXX := g++

DEBUG := -D_DEBUG  -DUSE_QLOG
DEBUG := -DNDEBUG  -DUSE_QLOG
#	-MMD -Wshadow -Wcast-qual -Wcast-align 
CXXFLAGS := -g3 $(DEBUG) -fPIC  \
	-MMD -Wcast-qual -Wcast-align \
	-Wwrite-strings -Wsign-compare  -Winvalid-pch -fms-extensions \
	-Wfloat-equal -Wextra -Wall -Wno-unused-parameter \
	-Wno-unused-function -Wunused-variable -Wunused-value \
	-Woverloaded-virtual -Wsign-promo -fno-gnu-keywords \
	-DH_USE_SYMC \
	-I.\
	-I/home/s/safe/include/protobuf3 \
	-I/home/s/safe/include \
	-I/home/s/include \
	-I/usr/local/include\


LDFLAGS := -L/home/s/safe/protobuf3/lib/\
		   -L/usr/local/lib \
	       -L/home/s/safe/lib  \
	       -L/home/s/lib 

#ifeq ($(PLATFORM), Linux)
#RPATH := -Wl,-rpath
#else
#RPATH := -rpath
#endif

RPATH := -Wl,-R


RTFLAGS := $(RPATH). \
	$(RPATH)/home/s/safe/protobuf3/lib/ \
	$(RPATH)/home/s/safe/lib \
	$(RPATH)/home/s/lib

LIBS := -lprotobuf\
	 /home/s/safe/lib/libboolexp.a \
	$(LIBMEMCACHED_LDFLAGS) \
	

PROTO_BASE:=/home/s/safe/bin/
PROTO:=protoc3
OUTPUT_DIR=$(shell pwd)
PROTOS := $(wildcard *.proto)
SRCS := $(wildcard *cgo.cc)  $(patsubst %.proto, ./%.pb.cc, $(PROTOS))
OBJS := $(patsubst %.cc, .obj/%.o, $(SRCS)) 
DEPS := $(patsubst %.o, %.d, $(OBJS))

MODULE_DIR  := $(shell dirname `pwd`)
#MODULE_NAME := $(shell basename $(MODULE_DIR))
MODULE_NAME := boolexp_cgo

TARGET := lib$(MODULE_NAME).so
TARGET_A_LIB := lib$(MODULE_NAME).a

MAJOR := 1
MINOR := 0.$(shell svnversion .)
FULL  := $(MAJOR).$(MINOR)
PKG   := backup/$(PLATFORM)_$(TARGET).$(FULL).tar.gz

BINPATH := 	/home/s/safe/$(MODULE_NAME)/lib
CONFPATH := /home/s/safe/$(MODULE_NAME)/etc
LOGPATH := 	/home/s/safe/$(MODULE_NAME)/logs/test
GOOGLE_PROTOBUF_VERSION:=3000001
all : prepare $(SRCS) 
	echo $(SRCS)
	echo $(PROTOS)
	make $(TARGET) 
	make $(TARGET_A_LIB)
	

$(TARGET) : $(OBJS) 
	$(CXX) $^ -Wl,-soname,$(@).$(MAJOR) -o $@.$(FULL) $(RTFLAGS) $(LDFLAGS) $(LIBS) -shared
	@rm -fr $(@)
	@rm -fr $(@).$(MAJOR)
	ln -s $(@).$(FULL) $(@).$(MAJOR)
	ln -s $(@).$(MAJOR) $(@)

$(TARGET_A_LIB): $(OBJS)
	@ar $(ARFLAGS)  $@ $^
	@ranlib $@   

clean:
	rm -rf .obj/*.o .obj/*.d $(TARGET)* *.so* *.a
	rm -rf *_pb2.py *.pb.cc *.pb.h *.pb.o *.pb.d *.proto.php  *.pb.go

check:
	go test

install : $(TARGET)
	#@mkdir -p backup
	#tar zcf $(PKG) $(TARGET) $(TARGET).$(MAJOR) $(TARGET).$(FULL)
	#@sudo -u cloud mkdir -p $(LOGPATH) $(CONFPATH) $(BINPATH)
	#@sudo -u cloud cp -r *.go  $(BINPATH)/../
	#@sudo -u cloud chmod a+w $(LOGPATH)
	#sudo -u cloud tar xvf $(PKG) -C $(BINPATH)
	#@sudo -u cloud cp ../proto/lib* $(BINPATH)
	#@sudo -u cloud cp  $(TARGET_A_LIB) $(BINPATH)

-include $(DEPS)

prepare : 
	@mkdir -p .obj  
	@mkdir -p .obj/

.obj/%.o : %.cc 
	$(CXX) -c $(CXXFLAGS) $< -o $@

%.pb.cc : %.proto
	$(PROTO_BASE)$(PROTO) --python_out=$(OUTPUT_DIR) --cpp_out=$(OUTPUT_DIR)  --go_out=$(OUTPUT_DIR) $< 


.PHONY: install check clean prepare


