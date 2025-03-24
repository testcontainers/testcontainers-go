#!/usr/bin/env bash

# exit on error, unset variables, print commands, fail on pipe errors
set -euxo pipefail

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
#    ALL_CHANGED_FILES="README.md" ./scripts/changed-modules.sh
#    The output should be: all modules.
#
# 6. A file from two modules in the modules dir are modified:
#    ALL_CHANGED_FILES="modules/nginx/go.mod modules/localstack/go.mod" ./scripts/changed-modules.sh
#    The output should be: the modules/nginx and modules/localstack modules.
#
# 7. Files from the excluded dirs are modified:
#    ALL_CHANGED_FILES="docs/a.md .vscode/a.json .devcontainer/a.json" ./scripts/changed-modules.sh
#    The output should be: no modules.
#
# 8. Several files in the same module are modified:
#    ALL_CHANGED_FILES="modules/nginx/go.mod modules/nginx/a.txt modules/nginx/b.txt" ./scripts/changed-modules.sh
#    The output should be: the modules/nginx module.
#
# 9. Several files in the core module are modified:
#    ALL_CHANGED_FILES="go.mod a.go b.go" ./scripts/changed-modules.sh
#    The output should be: all modules.
#
# 10. Several files in a build-excluded module are modified:
#    ALL_CHANGED_FILES="modules/k6/a.go" ./scripts/changed-modules.sh
#    The output should be: no modules.
#
# 11. Several files in different modules, including a build-excluded one, are modified:
#    ALL_CHANGED_FILES="modules/k6/a.go modules/clickhouse/a.txt" ./scripts/changed-modules.sh
#    The output should be: the modules/clickhouse module.
#
# 12. Several files in the core module are modified:
#    ALL_CHANGED_FILES="go.mod a.go b.go" ./scripts/changed-modules.sh
#    The output should be: all modules but the build-excluded ones.
#
# 13. A excluded module is modified with a file that is excluded:
#    ALL_CHANGED_FILES="modules/k6/a.go mkdocs.yml" ./scripts/changed-modules.sh
#    The output should be: no modules.
#
# 14. A excluded file and a file from the core module are modified:
#    ALL_CHANGED_FILES="mkdocs.yml go.mod" ./scripts/changed-modules.sh
#    The output should be: all modules.
#
# 15. Only excluded files are modified:
#    ALL_CHANGED_FILES="mkdocs.yml" ./scripts/changed-modules.sh
#    The output should be: no modules.
#
# There is room for improvement in this script. For example, it could detect if the changes applied to the docs or the .github dirs, and then do not include any module in the list.
# But then we would need to verify the CI scripts to ensure that the job receives the correct modules to build.

# ROOT_DIR is the root directory of the repository.
readonly ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)

# define an array of modules that won't be included in the list
readonly excluded_modules=(".devcontainer" ".vscode" "docs")

# define an array of files that won't be included in the list
readonly excluded_files=("mkdocs.yml" ".github/dependabot.yml" ".github/workflows/sonar-*.yml")

# define an array of modules that won't be part of the build
readonly no_build_modules=("modules/k6")

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
    # check if the file is in one of the excluded files
    for exclude_file in ${excluded_files[@]}; do
        if [[ $file == $exclude_file ]]; then
            # if the file is in the excluded files, skip the rest of the loop.
            # Execution continues at the loop control of the 2nd enclosing loop.
            continue 2
        fi
    done

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
        # check if the file is in one of the excluded modules
        for exclude_module in ${excluded_modules[@]}; do
            if [[ $file == $exclude_module/* ]]; then
                # continue skips to the next iteration of an enclosing for, select, until, or while loop in a shell script.
                # Execution continues at the loop control of the nth enclosing loop, in this case two levels up.
                continue 2
            fi
        done

        modified_modules=${allModules[@]}
        break
    fi
done

# print all modules with this format:
# each module will be enclosed in double quotes
# each module will be separated by a comma
# the entire list will be enclosed in square brackets
# the list will be sorted and unique
sorted_unique_modules=($(echo "${modified_modules[@]}" | tr ' ' '\n' | sort -u | tr '\n' ' '))

# remove modules that won't be part of the build from the list
filtered_modules=()
for module in "${sorted_unique_modules[@]}"; do
    skip=false
    for no_build_module in "${no_build_modules[@]}"; do
        if [[ ${module} == \"${no_build_module}\" ]]; then
            skip=true
            break
        fi
    done
    if [[ $skip == false ]]; then
        filtered_modules+=(${module})
    fi
done
sorted_unique_modules=("${filtered_modules[@]}")

echo "["$(IFS=,; echo "${sorted_unique_modules[*]}" | sed 's/ /,/g')"]"
