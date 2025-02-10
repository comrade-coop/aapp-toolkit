cmake_minimum_required(VERSION 3.14)
project(MinimalHttpServer)

# Use C++17.
set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED True)

include(FetchContent)

# Fetch nlohmann/json.
FetchContent_Declare(
    json
    GIT_REPOSITORY https://github.com/nlohmann/json.git
    GIT_TAG v3.11.2
)
FetchContent_MakeAvailable(json)

add_executable(server main.cpp)

# Link the header-only nlohmann/json library.
target_link_libraries(server PRIVATE nlohmann_json::nlohmann_json)