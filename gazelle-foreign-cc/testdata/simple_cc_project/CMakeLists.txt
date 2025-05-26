cmake_minimum_required(VERSION 3.10)
project(SimpleCCProject CXX)

# Define a library
add_library(my_lib lib.cc lib.h)

# Define an executable
add_executable(app main.cc)

# Link the library to the executable
target_link_libraries(app PRIVATE my_lib)
