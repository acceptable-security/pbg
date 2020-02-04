#!/bin/bash

bold=$(tput bold)
normal=$(tput sgr0)
red=$(tput setaf 1)

# Basic logging infrastructure
log_info() {
	echo     "${bold}INFO [$(date)] ${normal} $1"
}

log_error() {
	echo >&2 "${bold}ERR  [$(date)] ${normal} ${red}$1${normal}"
	exit 1
}

validate_cmd() {
	log_info "Checking $1 installation"

	{
		command -v $1 >/dev/null 2>&1
	} || {
		log_error "Please install $1 before running"
	}
}

validate_cmd "go"
validate_cmd "git"
validate_cmd "make"
validate_cmd "apt-get"

curr_dir=$(pwd)

log_info "Installing dependencies for dynamorio"
{
	sudo apt-get install cmake g++ g++-multilib doxygen transfig imagemagick ghostscript git zlib1g-dev
} || {
	log_error "Failed to install dependencies for dynamorio"
}

log_info "Cloning dynamorio to parent directory"
{
	git clone https://github.com/DynamoRIO/dynamorio.git ../dynamorio/
} || {
	log_error "Failed to clone dynamorio"
}

log_info "Building dynamorio"
{
	cd ../dynamorio && \
	mkdir ./build && \
	cd ./build && \
	cmake .. && \
	make -j && \
	DYNAMORIO_HOME=$(pwd) && \
	cd .. && \
	mkdir build_samples && \
	cd build_samples && \
	cmake -DDynamoRIO_DIR=$DYNAMORIO_HOME/cmake $DYNAMORIO_HOME/api/samples && \
	make
} || {
	log_error "Failed to build dynamorio"
}

log_info "Cloning capstone to the parent dictory"
{
	cd $curr_dir && \
	git clone https://github.com/aquynh/capstone.git ../capstone/
} || {
	log_error "Failed to clone capstone"
}

log_info "Installing capstone from parent directory"
{
	cd ../capstone/ && \
	sh ./make.sh && \
	sudo sh ./make.sh install
}

log_info "Building pbg"
{
	cd $curr_dir && \
	go build -o pbg
} || {
	log_error "Failed to build pbg"
}