cmake_minimum_required(VERSION 3.5)
project(PlatformTest)

# Configure a platform header file
configure_file(include/config.h.in config/platform.hpp)

add_library(platform_lib src/main.cpp)
target_include_directories(platform_lib PRIVATE config)