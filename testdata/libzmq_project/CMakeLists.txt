cmake_minimum_required(VERSION 3.10)
project(libzmq VERSION 4.3.4 LANGUAGES C CXX)

set(CMAKE_CXX_STANDARD 11)
set(CMAKE_C_STANDARD 99)

# Include directories
include_directories(include)
include_directories(src)

# Find required packages
find_package(Threads REQUIRED)

# Define the main libzmq library
add_library(zmq SHARED
    src/zmq.cpp
    src/socket.cpp
    src/context.cpp
    src/message.cpp
    src/poller.cpp
    src/address.cpp
    src/tcp_connecter.cpp
    src/tcp_listener.cpp
    include/zmq.h
    include/zmq_utils.h
)

# Set library properties
target_include_directories(zmq PUBLIC include)
target_include_directories(zmq PRIVATE src)

# Link with system libraries
target_link_libraries(zmq PRIVATE Threads::Threads)

# Optional: Create a static version
add_library(zmq-static STATIC
    src/zmq.cpp
    src/socket.cpp
    src/context.cpp
    src/message.cpp
    src/poller.cpp
    src/address.cpp
    src/tcp_connecter.cpp
    src/tcp_listener.cpp
)

target_include_directories(zmq-static PUBLIC include)
target_include_directories(zmq-static PRIVATE src)
target_link_libraries(zmq-static PRIVATE Threads::Threads)

# Create a simple test executable
add_executable(zmq_test
    tests/test_basic.cpp
)

target_link_libraries(zmq_test PRIVATE zmq)
target_include_directories(zmq_test PRIVATE include)

# Create a performance test
add_executable(perf_inproc_lat
    perf/inproc_lat.cpp
)

target_link_libraries(perf_inproc_lat PRIVATE zmq)