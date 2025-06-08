cmake_minimum_required(VERSION 3.10)
project(ConfigureFileExample VERSION 1.2.3)

# Set some variables
set(PROJECT_DESCRIPTION "A test project for configure_file")
set(ENABLE_FEATURE ON)

# Configure header file
configure_file(config.h.in config.h @ONLY)

# Create a library that uses the configured header
add_library(mylib src/lib.cpp)
target_include_directories(mylib PUBLIC ${CMAKE_CURRENT_BINARY_DIR})

# Create an executable
add_executable(app src/main.cpp)
target_link_libraries(app mylib)