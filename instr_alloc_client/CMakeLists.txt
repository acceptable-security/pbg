cmake_minimum_required(VERSION 3.10)
add_library(instr_alloc SHARED instr_alloc.c)

find_package(DynamoRIO)
if (NOT DynamoRIO_FOUND)
  message(FATAL_ERROR "DynamoRIO package required to build")
endif(NOT DynamoRIO_FOUND)

configure_DynamoRIO_client(instr_alloc)
use_DynamoRIO_extension(instr_alloc drwrap)
use_DynamoRIO_extension(instr_alloc drmgr)
use_DynamoRIO_extension(instr_alloc drreg)
