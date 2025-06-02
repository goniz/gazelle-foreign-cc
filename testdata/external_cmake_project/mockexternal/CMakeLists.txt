cmake_minimum_required(VERSION 3.10)
project(MockExternalLib)

add_library(external_lib src/external_lib.cpp)
target_include_directories(external_lib PUBLIC include)

add_executable(external_app src/main.cpp)
target_link_libraries(external_app external_lib)