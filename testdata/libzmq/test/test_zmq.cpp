#include <zmq.h>
#include <iostream>

int main() {
    void *context = zmq_ctx_new();
    if (context) {
        std::cout << "ZMQ context created successfully" << std::endl;
        zmq_ctx_destroy(context);
        return 0;
    }
    return 1;
}