#include <stdlib.h>
#include <stdio.h>
#include <sys/socket.h>
#include <arpa/inet.h>
#include <unistd.h>

int main(int argc, char** argv)
{
    printf("the writev example for https://github.com/golang/go/issues/13451\n");

    if (argc < 2) {
        printf("Usage: %s <port>\n"
            "   port: the tcp port to connect to.\n"
            "For example:\n"
            "   %s 1985\n", argv[0], argv[0]);
        exit(-1);
    }

    int port = ::atoi(argv[1]);
    printf("connect to server at tcp://%d\n", port);

    int fd = ::socket(AF_INET, SOCK_STREAM, 0);
    if (fd == -1) {
        printf("create socket failed.\n");
        exit(-1);
    }

    sockaddr_in addr;
    addr.sin_family = AF_INET;
    addr.sin_port = htons(port);
    addr.sin_addr.s_addr = inet_addr("127.0.0.1");

    if(::connect(fd, (const struct sockaddr*)&addr, sizeof(sockaddr_in)) < 0) {
        printf("connect server failed.\n");
        exit(-1);
    }

    // read util EOF.
    char buf[4096];
    while (true) {
        if (::read(fd, buf, 4096) == -1) {
            printf("server closed.\n");
            break;
        }
    }

    ::close(fd);

    return 0;
}
