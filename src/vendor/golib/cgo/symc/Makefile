# set CPUS for Linux or FreeBSD
PLATFORM := $(shell uname)
CPUS := $(strip $(if $(shell echo $(PLATFORM)|grep Linux),\
	$(shell cat /proc/cpuinfo|grep -c processor),\
	$(shell sysctl -a | egrep -i 'hw.ncpu' | cut -d: -f2)))


CC := gcc
CXX := g++

DEBUG := -DNDEBUG 

CXXFLAGS := -g3 $(DEBUG) -fPIC  \
	-MMD -Wcast-qual -Wcast-align \
	-Wwrite-strings -Wsign-compare -Werror -Winvalid-pch -fms-extensions \
	-Wfloat-equal -Wextra -Wall -Wno-unused-parameter \
	-Wno-unused-function -Wunused-variable -Wunused-value \
	-Woverloaded-virtual -Wsign-promo -fno-gnu-keywords \
	-DH_USE_SYMC \
	-I.\
	-I..\
	-I/home/s/safe/include \
	-I/home/s/include \
	-I/usr/local/include\


LDFLAGS := \
		   -L/usr/local/lib \
	       -L/home/s/safe/lib  \
	       -L/home/s/lib 

#ifeq ($(PLATFORM), Linux)
#RPATH := -Wl,-rpath
#else
#RPATH := -rpath
#endif

RPATH := -Wl,-R

#LIBMEMCACHED_LDFLAGS= -lsymcwrap -lsymc
LIBMEMCACHED_LDFLAGS= -lsymc
#LIBMEMCACHED_LDFLAGS=-lsymcwrap_wu -lmemcached

RTFLAGS := $(RPATH). \
	$(RPATH)/home/s/safe/lib \
	$(RPATH)/home/s/lib

LIBS := -lprotobuf\
	-lqoslib\
	$(LIBMEMCACHED_LDFLAGS) \
	

SRCS := $(wildcard *.cc)


OBJS := $(patsubst %.cc, .obj/%.o, $(SRCS))
DEPS := $(patsubst %.o, %.d, $(OBJS))

MODULE_DIR  := $(shell dirname `pwd`)
#MODULE_NAME := $(shell basename $(MODULE_DIR))
MODULE_NAME := symc_cgo

TARGET := lib$(MODULE_NAME).so
TARGET_A_LIB := lib$(MODULE_NAME).a

MAJOR := 1
MINOR := 0.$(shell svnversion .)
FULL  := $(MAJOR).$(MINOR)
PKG   := backup/$(PLATFORM)_$(TARGET).$(FULL).tar.gz

BINPATH := 	/home/s/safe/$(MODULE_NAME)
CONFPATH := /home/s/safe/$(MODULE_NAME)/etc
LOGPATH := 	/home/s/safe/$(MODULE_NAME)/logs/test
GOOGLE_PROTOBUF_VERSION:=3000001
all : prepare $(TARGET) $(TARGET_A_LIB)
	go build -x -v

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

check:
	echo ok
	go test 

install : $(TARGET)
	sudo -u cloud mkdir -p $(BINPATH)
	sudo -u cloud cp  -rf $(TARGET_A_LIB) $(BINPATH)
	sudo -u cloud cp  -rf symc.h $(BINPATH)

-include $(DEPS)

prepare : 
	@mkdir -p .obj  
	@mkdir -p .obj/

.obj/%.o : %.cc 
	$(CXX) -c $(CXXFLAGS) $< -o $@

.PHONY: install check clean prepare


