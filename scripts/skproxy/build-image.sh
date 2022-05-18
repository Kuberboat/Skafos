#!/bin/bash

parent_path=$(
	cd "$(dirname "${BASH_SOURCE[0]}")"
	pwd -P
)
source $parent_path/../common.sh
skproxy_obj=skproxy
build_image_dir=$assets_dir/skproxy
image_name=gun9nir/skproxy

# Copy executable to the same directory as Dockerfile.
cp $build_dir/$skproxy_obj $build_image_dir
cd $build_image_dir

# Build the image.
docker build -t $image_name .
ret_val=$?
rm $skproxy_obj
cd - &> /dev/null

if [ $ret_val -ne 0 ]
then
    return ret_val
fi

# Upload the image to Docker Hub.
# Credentials are required to push the image.
docker push $image_name:latest
