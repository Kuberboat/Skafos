#!/bin/bash

parent_path=$(
	cd "$(dirname "${BASH_SOURCE[0]}")"
	pwd -P
)
proj_root_path=$parent_path/..
build_dir=$proj_root_path/out/bin
assets_dir=$proj_root_path/assets
