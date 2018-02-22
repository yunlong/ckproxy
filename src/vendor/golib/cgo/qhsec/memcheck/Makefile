
CXX      := g++
CXXFLAGS := -Wall -Werror -g \
	-I/home/s/safe/include \
	-I/home/s/include \
	-I..

LDFLAGS := -L/home/s/safe/lib -L. -lnetproto 
RTFLAGS := -Wl,-rpath=/home/s/safe/lib \
	-Wl,-rpath=/home/s/lib

all : sec
	valgrind --trace-children=yes --leak-check=full ./sec

sec : sec.cc qhsec.o
	$(CXX) $(CXXFLAGS) $^ -o $@ $(LDFLAGS) $(RTFLAGS)

qhsec.o : ../qhsec.cc
	$(CXX) -c $(CXXFLAGS) $< -o $@

clean :
	-rm -f ./sec ./*.o

.PHONY: all clean
