CMAKE_MINIMUM_REQUIRED(VERSION 3.5)
PROJECT(kungfu)

SET(CMAKE_CXX_STANDARD 14)

ADD_DEFINITIONS(-Wfatal-errors)
ADD_DEFINITIONS(-Wall)
ADD_COMPILE_OPTIONS(-fPIC)

SET(CMAKE_MODULE_PATH ${CMAKE_SOURCE_DIR}/cmake)

OPTION(KUNGFU_BUILD_TESTS "Build tests." OFF)
OPTION(KUNGFU_BUILD_TF_OPS "Build tensorflow operators." OFF)
OPTION(KUNGFU_BUILD_TOOLS "Build kungfu tools." OFF)
OPTION(KUNGFU_ENABLE_FLOAT16 "Enable float16." ON)
OPTION(KUNGFU_ENABLE_TRACE "Enable trace." OFF)

INCLUDE(cmake/go.cmake)
IF(KUNGFU_ENABLE_TRACE)
    ADD_DEFINITIONS(-DKUNGFU_ENABLE_TRACE)
    FIND_PACKAGE(stdtracer REQUIRED)
ENDIF()

INCLUDE_DIRECTORIES(${CMAKE_SOURCE_DIR}/srcs/cpp/include)
LINK_DIRECTORIES(${LIBRARY_OUTPUT_PATH})

IF(KUNGFU_ENABLE_FLOAT16)
    SET(CGO_CFLAGS "-DENABLE_F16 -mf16c -mavx")
ENDIF()

INCLUDE(cmake/cgo.cmake)
ADD_CGO_LIBRARY(libkungfu-comm ${CMAKE_SOURCE_DIR}/srcs/go/libkungfu-comm)

ADD_LIBRARY(kungfu SHARED srcs/cpp/src/kungfu.cpp)
TARGET_LINK_LIBRARIES(kungfu kungfu-comm)
ADD_DEPENDENCIES(kungfu libkungfu-comm)
ADD_CGO_DEPS(kungfu)

INSTALL(DIRECTORY srcs/cpp/include DESTINATION .)
INSTALL(DIRECTORY lib DESTINATION .)

IF(KUNGFU_ENABLE_NCCL)
    INCLUDE(cmake/nccl.cmake)
ENDIF()

IF(KUNGFU_BUILD_TF_OPS)
    INCLUDE(cmake/tf-op.cmake)

    ADD_LIBRARY(kungfu_python SHARED srcs/cpp/src/python/init.cpp
                                     srcs/cpp/src/python/cuda.cpp)
    TARGET_LINK_LIBRARIES(kungfu_python kungfu)

    FILE(GLOB CPU_OP_SRCS
         ${CMAKE_SOURCE_DIR}/srcs/cpp/src/tensorflow/ops/cpu/*.cpp)
    FILE(GLOB GPU_OP_SRCS
         ${CMAKE_SOURCE_DIR}/srcs/cpp/src/tensorflow/ops/gpu/*.cpp)

    ADD_TF_OP_LIB(kungfu_tensorflow_ops ${CPU_OP_SRCS})
    TARGET_LINK_LIBRARIES(kungfu_tensorflow_ops kungfu_python kungfu)

    IF(KUNGFU_ENABLE_NCCL)
        TARGET_SOURCES(kungfu_python PRIVATE srcs/cpp/src/python/nccl.cpp
                       srcs/cpp/src/nccl/gpu_collective.cpp)
        USE_NCCL(kungfu_python)
        TARGET_SOURCES(kungfu_tensorflow_ops PRIVATE ${GPU_OP_SRCS})
    ENDIF()
    IF(KUNGFU_ENABLE_TRACE)
        USE_STDTRACER(kungfu_python)
        USE_STDTRACER(kungfu_tensorflow_ops)
    ENDIF()

    SET_TF_COMPILE_OPTION(kungfu_python)
ENDIF()

IF(KUNGFU_BUILD_TOOLS)
    FUNCTION(ADD_GO_BINARY target)
        FILE(MAKE_DIRECTORY ${CMAKE_RUNTIME_OUTPUT_DIRECTORY})
        ADD_CUSTOM_TARGET(
            ${target} ALL
            COMMAND
                env #
                GOBIN=${CMAKE_RUNTIME_OUTPUT_DIRECTORY}
                CGO_CFLAGS=${CGO_CFLAGS} CGO_LDFLAGS=${CGO_LDFLAGS}
                CGO_CXXFLAGS=${CGO_CXXFLAGS} ${GO} install -v ${ARGN})
    ENDFUNCTION()

    FUNCTION(ADD_KUNGFU_GO_BINARY target)
        ADD_GO_BINARY(${target} ${CMAKE_SOURCE_DIR}/srcs/go/cmd/${target})
    ENDFUNCTION()

    ADD_KUNGFU_GO_BINARY(kungfu-run)
ENDIF()

IF(KUNGFU_BUILD_TESTS)
    ENABLE_TESTING()
    INCLUDE(cmake/tests.cmake)
ENDIF()
