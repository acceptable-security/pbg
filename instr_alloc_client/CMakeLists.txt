cmake_minimum_required(VERSION 3.10)
add_library(myclient SHARED myclient.c)

find_package(DynamoRIO)
if (NOT DynamoRIO_FOUND)
  message(FATAL_ERROR "DynamoRIO package required to build")
endif(NOT DynamoRIO_FOUND)

configure_DynamoRIO_client(myclient)
use_DynamoRIO_extension(myclient drwrap)
use_DynamoRIO_extension(myclient drmgr)
