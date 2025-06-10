#include <iostream>
#include "config.h"

int main() {
    std::cout << "Version: " << MY_VERSION << std::endl;
    std::cout << "Feature enabled: " << FEATURE_ENABLED << std::endl;
    return 0;
}