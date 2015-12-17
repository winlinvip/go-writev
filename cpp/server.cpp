#include <stdlib.h>
#include <stdio.h>
#include <sys/socket.h>
#include <arpa/inet.h>
#include <unistd.h>
#include <sys/uio.h>
#include <string.h>

#include <string>
using namespace std;

#define NB_VIDEOS_IN_GROUP 512
#define VIDEO_SIZE 4096
#define HEADER_SIZE 12

int srs_send(int fd, char** group, bool use_writev, bool write_one_by_one);

int main(int argc, char** argv)
{
    printf("the writev example for https://github.com/golang/go/issues/13451\n");

    int port = 0;
    bool use_writev = false, write_one_by_one = false;
    if (argc >= 3) {
        port = ::atoi(argv[1]);
        use_writev = string(argv[2]) == "true";
        write_one_by_one = false;
    }
    if (!use_writev && argc >= 4) {
        write_one_by_one = string(argv[3]) == "true";
    }
    if (argc < 3 || (!use_writev && argc < 4)) {
        printf("Usage: %s <port> <use_writev> [write_one_by_one]\n"
            "   port: the tcp listen port.\n"
            "   use_writev: whether use writev. true or false.\n"
            "   write_one_by_one: for write(not writev), whether send packet one by one.\n"
            "Fox example:\n"
            "   %s 1985 true\n"
            "   %s 1985 false true\n"
            "   %s 1985 false false\n", argv[0], argv[0], argv[0], argv[0]);
        exit(-1);
    }

    printf("listen at tcp://%d, use %s\n", port, (use_writev? "writev":"write"));
    if (!use_writev) {
        printf("for write, send %s\n", (write_one_by_one? "one-by-one":"copy-to-big-buffer"));
    }

    int fd = 0;
    if ((fd = socket(AF_INET, SOCK_STREAM, 0)) == -1) {
        printf("create linux socket error.\n");
        exit(-1);
    }

    int reuse_socket = 1;
    if (::setsockopt(fd, SOL_SOCKET, SO_REUSEADDR, &reuse_socket, sizeof(int)) == -1) {
        printf("setsockopt reuse-addr error.\n");
        exit(-1);
    }

    sockaddr_in addr;
    addr.sin_family = AF_INET;
    addr.sin_port = htons(port);
    addr.sin_addr.s_addr = inet_addr("0.0.0.0");
    if (::bind(fd, (const sockaddr*)&addr, sizeof(sockaddr_in)) == -1) {
        printf("bind socket error.\n");
        exit(-1);
    }

    if (::listen(fd, 10) == -1) {
        printf("listen socket error.\n");
        exit(-1);
    }

    while (true) {
        int client = ::accept(fd, NULL, NULL);
        if (client == -1) {
            printf("accept failed.\n");
            exit(-1);
        }

        // assume there is a video stream, which contains infinite video packets,
        // server must delivery all video packets to client.
        // for high performance, we send a group of video(to avoid too many syscall),
        // here we send 10 videos as a group.
        while (true) {
            // @remark for test, each video is M bytes.
            char* video = new char[VIDEO_SIZE];

            // @remark for test, each video header is M0 bytes.
            char* header = new char[HEADER_SIZE];

            // @remark for test, each group contains N (header+video)s.
            char** group = new char*[2 * NB_VIDEOS_IN_GROUP];
            for (int i = 0; i < 2 * NB_VIDEOS_IN_GROUP; i+= 2) {
                group[i] = header;
                group[i + 1] = video;
            }

            // sendout the video group.
            int sent = srs_send(client, group, use_writev, write_one_by_one);
            delete[] video;
            delete[] group;
            delete[] header;
            if (sent == -1) {
                printf("sendout the video group failed, client closed?\n");
                ::close(client);
                break;
            }
        }
    }

    return 0;
}

// each group contains N (header+video)s.
//      header is M bytes.
//      videos is M0 bytes.
int srs_send(int fd, char** group, bool use_writev, bool write_one_by_one)
{
    if (use_writev) {
        iovec iovs[2 * NB_VIDEOS_IN_GROUP];
        for (int i = 0; i < 2 * NB_VIDEOS_IN_GROUP; i+=2) {
            iovs[i].iov_base = (char*)group[i];
            iovs[i].iov_len = HEADER_SIZE;

            iovs[i+1].iov_base = (char*)group[i + 1];
            iovs[i+1].iov_len = VIDEO_SIZE;
        }

        return writev(fd, iovs, 2 * NB_VIDEOS_IN_GROUP);
    }

    // use write, send one by one packet.
    // @remark avoid memory copy, but with lots of syscall, hurts performance.
    if (write_one_by_one) {
        for (int i = 0; i < 2 * NB_VIDEOS_IN_GROUP; i+=2) {
            if (::write(fd, group[i], HEADER_SIZE) == -1) {
                return -1;
            }

            if (::write(fd, group[i + 1], VIDEO_SIZE) == -1) {
                return -1;
            }
        }
        return 0;
    }

    // use write, to avoid lots of syscall, we copy to a big buffer.
    char* buf = new char[NB_VIDEOS_IN_GROUP * (HEADER_SIZE + VIDEO_SIZE)];

    int nn = 0;
    for (int i = 0; i < 2 * NB_VIDEOS_IN_GROUP; i+=2) {
        memcpy(buf + nn, group[i], HEADER_SIZE);
        nn += HEADER_SIZE;

        memcpy(buf + nn, group[i + 1], VIDEO_SIZE);
        nn += VIDEO_SIZE;
    }

    nn = ::write(fd, buf, nn);
    delete[] buf;
    return nn;
}
