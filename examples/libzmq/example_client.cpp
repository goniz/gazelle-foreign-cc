#include <zmq.h>
#include <iostream>
#include <string>
#include <chrono>
#include <thread>

int main() {
    // Initialize ZMQ context
    void* context = zmq_ctx_new();
    
    // Create a socket
    void* socket = zmq_socket(context, ZMQ_REQ);
    
    // Connect to server
    int rc = zmq_connect(socket, "tcp://localhost:5555");
    if (rc != 0) {
        std::cerr << "Failed to connect: " << zmq_strerror(zmq_errno()) << std::endl;
        return 1;
    }
    
    std::cout << "ZMQ Client connected to server" << std::endl;
    
    // Send requests
    for (int i = 0; i < 3; ++i) {
        std::string request = "Hello from client " + std::to_string(i);
        
        // Send message
        zmq_send(socket, request.c_str(), request.length(), 0);
        std::cout << "Sent: " << request << std::endl;
        
        // Receive reply
        char buffer[256];
        int size = zmq_recv(socket, buffer, 255, 0);
        if (size != -1) {
            buffer[size] = '\0';
            std::cout << "Received: " << buffer << std::endl;
        }
        
        // Small delay
        std::this_thread::sleep_for(std::chrono::milliseconds(100));
    }
    
    // Cleanup
    zmq_close(socket);
    zmq_ctx_destroy(context);
    
    std::cout << "Client finished" << std::endl;
    return 0;
}