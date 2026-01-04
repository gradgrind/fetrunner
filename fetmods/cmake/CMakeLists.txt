cmake_minimum_required(VERSION 3.20)
project(fet VERSION 7.6.4 LANGUAGES C CXX)
set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

find_package(Qt6 6.3 REQUIRED COMPONENTS Core)

# Set a default build type if none is specified by the user
if(NOT CMAKE_BUILD_TYPE)
  set(CMAKE_BUILD_TYPE Release CACHE STRING "Choose the build type from available options" FORCE)
  # Set the available options for the user
  set_property(CACHE CMAKE_BUILD_TYPE PROPERTY STRINGS "Debug" "Release" "RelWithDebInfo" "MinSizeRel")
endif()

if(INSTALL_BASE)
  cmake_path(APPEND CMAKE_BINARY_DIR ${INSTALL_BASE} OUTPUT_VARIABLE installpath)
  set(CMAKE_INSTALL_PREFIX ${installpath})
endif()

MESSAGE(STATUS "Using toolchain file: ${CMAKE_TOOLCHAIN_FILE}")

if(WIN32)
add_subdirectory(src-cl)
endif()
#add_subdirectory(src)
add_subdirectory(fetrunner/gui)

MESSAGE(STATUS "Using install prefix: ${CMAKE_INSTALL_PREFIX}")
