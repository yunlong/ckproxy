CXXFLAGS:= -I/home/s/safe/include -I/home/s/include 
LDFLAGS:= -L/home/s/safe/lib -L/home/s/lib
LDFLAGS+= -lstdc++ /home/s/safe/lib/libnetproto.a /home/s/safe/lib/libcryptopp.a /home/s/safe/lib/libnppnacl.a /home/s/safe/lib/libqoslib.a -lcrypto -lz -lpthread
LDFLAGS+= -Wl,-rpath=/home/s/safe/lib
LDFLAGS+= -Wl,-rpath=/home/s/lib

BIN = qhsec_test

$(BIN) : main.cc
	g++ -g main.cc $(CXXFLAGS) $(LDFLAGS) -o $@

clean :
	rm -rf $(BIN)
