cmake_minimum_required(VERSION 3.10)
project(SimpleLib)

# Define a simple library
add_library(simple_lib 
    src/simple.cpp
    src/utils.cpp
)

target_include_directories(simple_lib PUBLIC include)

# Define a simple executable
add_executable(simple_app src/main.cpp)
target_link_libraries(simple_app simple_lib)