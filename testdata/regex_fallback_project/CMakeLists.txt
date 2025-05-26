cmake_minimum_required(VERSION 3.10)
project(RegexFallbackTest)

add_executable(simple_app main.cpp utils.cpp)
add_library(simple_lib STATIC helper.cpp)