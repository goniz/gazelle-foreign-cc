# CMake build script for ZeroMQ
project(ZeroMQ)

cmake_minimum_required(VERSION 2.8.12)

# Set lists to empty beforehand as to not accidentally take values from parent
set(sources)
set(cxx-sources)

# -----------------------------------------------------------------------------
# source files

set(cxx-sources
    address.cpp
    channel.cpp
    client.cpp
    clock.cpp
    ctx.cpp
    zmq.cpp
    # headers
    address.hpp
    channel.hpp
    client.hpp
    clock.hpp
    ctx.hpp
)

# source generators
foreach(source ${cxx-sources})
  list(APPEND sources ${CMAKE_CURRENT_SOURCE_DIR}/src/${source})
endforeach()

set(public_headers
    include/zmq.h
    include/zmq_utils.h
)

# -----------------------------------------------------------------------------
# output

# Main shared library
add_library(libzmq SHARED ${sources} ${public_headers})

set_target_properties(libzmq PROPERTIES
  PUBLIC_HEADER "${public_headers}"
  COMPILE_DEFINITIONS "DLL_EXPORT")

# Static library
add_library(libzmq-static STATIC ${sources} ${public_headers})

set_target_properties(libzmq-static PROPERTIES
  PUBLIC_HEADER "${public_headers}"
  COMPILE_DEFINITIONS "ZMQ_STATIC")

# Simple test executable
add_executable(test_zmq test/test_zmq.cpp)
target_link_libraries(test_zmq libzmq)