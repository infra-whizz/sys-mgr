#!/bin/bash

# Configuration
P_NAME="sysroot-manager"
P_APP="cmd/sys-mgr.go"
P_SRC_DIRS=("arch" "pm" "pm/fixlets" "sr" "cmd")
P_DOC_DIRS=("etc")
P_FILES=("LICENSE" "README.md" "go.mod" "go.sum" "utils.go" "sysmgr.go")
P_CMD=("Makefile")

set -e

#
# Check if directory exist
#
function dir_exists {
    if [ ! -d $1 ]; then
	echo "ERROR: Directory $1 was not found."
	if [ "$1" != "" ]; then
	    echo "Hint: $1"
	fi
	exit 1
    fi
}


#
# Check correct location of the script launch
#
function check_location {
    c_path=$(pwd)
    src_path=$(dirname "$(readlink -f "$0")")
    if [ "$src_path" != "$c_path" ]; then
       echo "This script should be ran from the same directory where it is"
       exit 1
    fi
}

#
# Get current version of the repodiff
#
function get_version {
    echo $(cat ../$P_APP | awk '/var VERSION/ {split($0,v,"\""); print v[2]}')
}

#
# Prepare space for the data content
#
function prepare_space {
    d_name="$P_NAME-$(get_version)"
    rm -rf $d_name > /dev/null
    mkdir $d_name
    echo $d_name
}

#
# Copy everything that is going to be a package
#
function copy_packaged_sources {
    dst=$1
    for d in ${P_SRC_DIRS[@]}; do
	echo "Copying source directory $d to $dst..."
	mkdir -p $dst/$d
	cp -r ../$d/*.go $dst/$d
    done

    for d in ${P_DOC_DIRS[@]}; do
	echo "Copying documentation directory $d to $dst..."
	mkdir -p $dst/$d
	cp -r ../$d $dst
    done

    # copy cmd
    for f in ${P_CMD[@]}; do
	echo "Copying $f to $dst/cmd..."
	cp ../cmd/$f $dst/cmd
    done

    # other
    for m in ${P_FILES[@]}; do
	echo "Copying $m file to the $dst..."
	cp ../$m $dst/
    done
}

function copy_vendor_sources {
    # copy vendor
    echo "Vendoring deps..."
    pushd ..
    go mod tidy
    go mod vendor
    popd

    v_dir="../vendor"
    dir_exists "$v_dir" "Please run 'go mod vendor' to make it."
    echo "Copying vendor libraries..."
    cp -r $v_dir .
}

#
# Create archive
#
function create_src_archive {
    dst=$1

    arc_name="$dst.tar.gz"
    dir_exists $dst "Permissions problem?"
    echo "Creating source archive..."
    tar cf - $dst | gzip -9 > $dst.tar.gz
}

function create_vendor_archive {
    arc_name="vendor.tar.gz"
    dir_exists "vendor" "No vendor directory has been found"
    echo "Creating vendor archive..."
    tar cf - vendor | gzip -9 > vendor.tar.gz
}


#
# Cleanup
#
function cleanup {
    dst=$1
    if [ -d $dst ]; then
	echo "Cleaning up temporary source..."
	rm -rf $dst
    fi
    if [ -d vendor ]; then
	echo "Cleaning up vendor..."
	rm -rf vendor
	rm -rf ../vendor
    fi
}


check_location
space=$(prepare_space)
copy_packaged_sources $space
create_src_archive $space
copy_vendor_sources
create_vendor_archive
cleanup $space
echo "Finished"
