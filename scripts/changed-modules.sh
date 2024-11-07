#!/usr/bin/env bash

# How to test this script, run it with the required environment variables:
# 1. A Go file from the core module is modified:
#    ALL_CHANGED_FILES="examples/nginx/go.mod examples/foo/a.txt a/b/c/d/a.go" ./scripts/changed-modules.sh
#    The output should be: all modules.
#
# 2. A file from a module in the modules dir is modified:
#    ALL_CHANGED_FILES="modules/nginx/go.mod" ./scripts/changed-modules.sh
#    The output should be: just the modules/nginx module.
#
# 3. A file from a module in the examples dir is modified:
#    ALL_CHANGED_FILES="examples/nginx/go.mod" ./scripts/changed-modules.sh
#    The output should be: just the examples/nginx module.
#
# 4. A Go file from the modulegen dir is modified:
#    ALL_CHANGED_FILES="modulegen/a.go" ./scripts/changed-modules.sh
#    The output should be: just the modulegen module.
#
# 5. A non-Go file from the core dir is modified:
#    ALL_CHANGED_FILES="docs/README.md" ./scripts/changed-modules.sh
#    The output should be: all modules.
#
# 6. A file from two modules in the modules dir are modified:
#    ALL_CHANGED_FILES="modules/nginx/go.mod modules/localstack/go.mod" ./scripts/changed-modules.sh
#    The output should be: the modules/nginx and modules/localstack modules.
#
# There is room for improvement in this script. For example, it could detect if the changes applied to the docs or the .github dirs, and then do not include any module in the list.
# But then we would need to verify the CI scripts to ensure that the job receives the correct modules to build.

# ROOT_DIR is the root directory of the repository.
readonly ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)

# modules is an array that will store the paths of all the modules in the repository.
modules=()

# Find all go.mod files in the repository, building a list of all the available modules and examples.
for modFile in $(find "${ROOT_DIR}/modules" -name "go.mod" -not -path "${ROOT_DIR}/**/testdata/*"); do
    modules+=("\"modules/$(basename "$(dirname "${modFile}")")\"")
done
for modFile in $(find "${ROOT_DIR}/examples" -name "go.mod" -not -path "${ROOT_DIR}/**/testdata/*"); do
    modules+=("\"examples/$(basename "$(dirname "${modFile}")")\"")
done

# sort modules array
IFS=$'\n' modules=($(sort <<<"${modules[*]}"))
unset IFS

# capture the root module
readonly rootModule="\"\""

# capture the modulegen module
readonly modulegenModule="\"modulegen\""

# merge all modules and examples into a single array
allModules=(${rootModule} ${modulegenModule} "${modules[@]}")

# sort allModules array
IFS=$'\n' allModules=($(sort <<<"${allModules[*]}"))
unset IFS

# Get the list of modified files, retrieved from the environment variable ALL_CHANGED_FILES.
# On CI, this value will come from a Github Action retrieving the list of modified files from the pull request.
readonly modified_files=${ALL_CHANGED_FILES[@]}

# Initialize variables
modified_modules=()

# Check the modified files and determine which modules to build, following these rules:
# - if the modified files contain any file in the root module, include all modules in the list
# - if the modified files only contain files in one of the modules, include that module in the list
# - if the modified files only contain files in one of the examples, include that example in the list
# - if the modified files only contain files in the modulegen module, include only the modulegen module in the list
for file in $modified_files; do
    if [[ $file == modules/* ]]; then
        module_name=$(echo $file | cut -d'/' -f2)
        if [[ ! " ${modified_modules[@]} " =~ " ${module_name} " ]]; then
            modified_modules+=("\"modules/$module_name\"")
        fi
    elif [[ $file == examples/* ]]; then
        example_name=$(echo $file | cut -d'/' -f2)
        if [[ ! " ${modified_modules[@]} " =~ " ${example_name} " ]]; then
            modified_modules+=("\"examples/$example_name\"")
        fi
    elif [[ $file == modulegen/* ]]; then
        modified_modules+=("\"modulegen\"")
    else
        # a file from the core module is modified, so include all modules in the list and stop the loop
        modified_modules=${allModules[@]}
        break
    fi
done

# print all modules with this format:
# each module will be enclosed in double quotes
# each module will be separated by a comma
# the entire list will be enclosed in square brackets
echo "["$(IFS=,; echo "${modified_modules[*]}" | sed 's/ /,/g')"]"
