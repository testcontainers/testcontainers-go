#!/usr/bin/env bash

# Find all go.mod files in the repository, building a list of all the available modules.

# ROOT_DIR is the root directory of the repository.
readonly ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)

# modules is an array that will store the paths of all the modules in the repository.
modules=()

# capture modules
for modFile in $(find "${ROOT_DIR}/modules" -name "go.mod" -not -path "${ROOT_DIR}/**/testdata/*"); do
    modules+=("\"modules/$(basename "$(dirname "${modFile}")")\"")
done

# capture examples
for modFile in $(find "${ROOT_DIR}/examples" -name "go.mod" -not -path "${ROOT_DIR}/**/testdata/*"); do
    modules+=("\"examples/$(basename "$(dirname "${modFile}")")\"")
done

# sort modules array
IFS=$'\n' modules=($(sort <<<"${modules[*]}"))
unset IFS

# capture the root module
rootModule="\"\""

# capture the modulegen module
modulegenModule="\"modulegen\""

# merge all modules and examples into a single array
allModules=("${rootModule}" "${modulegenModule}" "${modules[@]}")

# sort allModules array
IFS=$'\n' allModules=($(sort <<<"${allModules[*]}"))

# print all modules with this format:
# each module will be enclosed in double quotes
# each module will be separated by a comma
# the entire list will be enclosed in square brackets
echo "$(IFS=,; echo "${allModules[*]}" | sed 's/ /,/g')"
