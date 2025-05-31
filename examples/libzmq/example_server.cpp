#include <zmq.h>
#include <iostream>
#include <string>
#include <chrono>
#include <thread>

int main() {
    // Initialize ZMQ context
    void* context = zmq_ctx_new();
    
    // Create a socket
    void* socket = zmq_socket(context, ZMQ_REP);
    
    // Bind to port
    int rc = zmq_bind(socket, "tcp://*:5555");
    if (rc != 0) {
        std::cerr << "Failed to bind: " << zmq_strerror(zmq_errno()) << std::endl;
        return 1;
    }
    
    std::cout << "ZMQ Server listening on port 5555" << std::endl;
    
    // Handle requests
    for (int i = 0; i < 3; ++i) {
        char buffer[256];
        
        // Receive request
        int size = zmq_recv(socket, buffer, 255, 0);
        if (size != -1) {
            buffer[size] = '\0';
            std::cout << "Received: " << buffer << std::endl;
            
            // Send reply
            std::string reply = "Echo: " + std::string(buffer);
            zmq_send(socket, reply.c_str(), reply.length(), 0);
            std::cout << "Sent: " << reply << std::endl;
        }
    }
    
    // Cleanup
    zmq_close(socket);
    zmq_ctx_destroy(context);
    
    std::cout << "Server finished" << std::endl;
    return 0;
}