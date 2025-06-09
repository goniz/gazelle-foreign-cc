#include "config.h"
#include <iostream>

void printConfig() {
    std::cout << "Project: " << PROJECT_NAME << std::endl;
    std::cout << "Version: " << PROJECT_VERSION << std::endl;
    std::cout << "Description: " << PROJECT_DESCRIPTION << std::endl;
#ifdef ENABLE_FEATURE
    std::cout << "Feature enabled" << std::endl;
#else
    std::cout << "Feature disabled" << std::endl;
#endif
}