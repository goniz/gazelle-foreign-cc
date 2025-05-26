cmake_minimum_required(VERSION 3.10)
project(ComplexProject)

set(CMAKE_CXX_STANDARD 17)

# Add include directories
include_directories(include)
include_directories(third_party/include)

# Create a static library
add_library(utils STATIC
    src/utils.cpp
    src/helper.cpp
)

# Create a shared library with dependencies
add_library(core SHARED
    src/core.cpp
    src/manager.cpp
)

target_link_libraries(core utils)
target_include_directories(core PUBLIC include)

# Create an executable that depends on both libraries
add_executable(main_app
    src/main.cpp
)

target_link_libraries(main_app core utils)

# Create a test executable
add_executable(test_runner
    tests/test_main.cpp
    tests/test_utils.cpp
)

target_link_libraries(test_runner utils)
target_include_directories(test_runner PRIVATE tests)

# Add subdirectory with its own CMakeLists.txt
add_subdirectory(plugins)