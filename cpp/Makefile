default: server client
server: server.o
	g++ server.o -o server
server.o: server.cpp Makefile
	g++ -g -O0 -Wall -c server.cpp -o server.o
client: client.o
	g++ client.o -o client
client.o: client.cpp Makefile
	g++ -g -O0 -Wall -c client.cpp -o client.o
