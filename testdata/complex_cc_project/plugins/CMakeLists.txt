add_library(plugin_system STATIC
    plugin.cpp
    loader.cpp
)

target_include_directories(plugin_system PUBLIC ../include)