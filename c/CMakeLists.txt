cmake_minimum_required(VERSION 2.8)

project(container-test)

aux_source_directory(src DIR_SRCS)

include_directories(include)

add_executable(container-test ${DIR_SRCS})

target_link_libraries(container-test cap)

set_target_properties(container-test PROPERTIES COMPILE_FLAGS "-Wall -g -c")

set(CMAKE_EXPORT_COMPILE_COMMANDS ON)
