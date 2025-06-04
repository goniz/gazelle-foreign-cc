#include "simple.h"
#include <iostream>

void SimpleLib::hello() {
    std::cout << "Hello from Simple Library!" << std::endl;
}

int SimpleLib::add(int a, int b) {
    return a + b;
}