cmake_minimum_required(VERSION 2.8)

project(container-test)

# set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/build)
# set(CMAKE_LIBRARY_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/build)
# set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/build)

aux_source_directory(src DIR_SRCS)

include_directories(include)

add_executable(container-test ${DIR_SRCS})

target_link_libraries(container-test cap)

set_target_properties(container-test PROPERTIES COMPILE_FLAGS "-Wall -g -c")

set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

set(BASESYS_PATH /home/solomon/workspace/basesystem/ubuntu_xenial_1604/)

enable_testing()

add_test(test container-test
    -h "conatiner"
    -p ${BASESYS_PATH}
    -m "./sh/network.sh:usr/local/bin/network.sh"
    -e "PATH=/usr/bin:/usr/local/bin:/usr/local/sbin:/bin:/sbin"
    )

# set_tests_properties (test
#     PROPERTIES PASS_REGULAR_EXPRESSION "ok")
