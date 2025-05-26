cmake_minimum_required(VERSION 3.10)
project(InvalidProject)

# This will cause CMake to fail
invalid_cmake_command_that_does_not_exist()

add_executable(app main.cpp)