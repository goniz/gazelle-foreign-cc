cmake_minimum_required(VERSION 3.10)
project(configure_file_test)

configure_file(config.h.in config.h @ONLY)