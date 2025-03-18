#!/bin/bash

# Prevent command echoing and exit on any error
set +x -e

# Clear sensitive environment variables on exit
trap 'unset SONAR_TOKEN' EXIT

# sonar-manager.sh
# ---------------
# Manages SonarCloud projects for the testcontainers-go repository
#
# Usage:
#   ./.github/scripts/sonar-manager.sh -h                         # Show help message
#   ./.github/scripts/sonar-manager.sh -a create -p <project>     # Create a single project
#   ./.github/scripts/sonar-manager.sh -a delete -p <project>     # Delete a single project
#   ./.github/scripts/sonar-manager.sh -a createAll               # Create all projects
#   ./.github/scripts/sonar-manager.sh -a deleteAll               # Delete all projects
#
# Project name format:
#   - For modules: modules_<module-name>
#   - For examples: examples_<example-name>
#   - For modulegen: modulegen
#
# Examples:
#   ./.github/scripts/sonar-manager.sh -a create -p modulegen
#   ./.github/scripts/sonar-manager.sh -a create -p modules_mysql
#   ./.github/scripts/sonar-manager.sh -a create -p examples_redis
#   ./.github/scripts/sonar-manager.sh -a delete -p modules_mysql
#   ./.github/scripts/sonar-manager.sh -a createAll
#   ./.github/scripts/sonar-manager.sh -a deleteAll
#
# Environment variables:
#   SONAR_TOKEN - Required. The SonarCloud authentication token.

SONAR_TOKEN=${SONAR_TOKEN:-}
ORGANIZATION="testcontainers"

# list all directories in the modules and examples directories
EXAMPLES=$(ls -1d examples/*/ | cut -d'/' -f2)
MODULES=$(ls -1d modules/*/ | cut -d'/' -f2)

# Delete all the projects in SonarCloud, starting with the modulegen module.
# Modules and examples need a prefix with the module type (module or example).
deleteAll() {
    delete "modulegen"

    for MODULE in $MODULES; do
        delete "modules_$MODULE"
    done

    for EXAMPLE in $EXAMPLES; do
        delete "examples_$EXAMPLE"
    done
}

# Helper function to print success message
print_success() {
    local action=$1
    local module=$2
    echo -e "\033[32mOK\033[0m: $action $([ ! -z "$module" ] && echo "for $module")"
}

# Helper function to print failure message
print_failure() {
    local action=$1
    local module=$2
    local status=$3
    echo -e "\033[31mFAIL\033[0m: Failed to $action $([ ! -z "$module" ] && echo "for $module") (HTTP status: $status)"
    exit 1
}

# Helper function to handle curl responses
handle_curl_response() {
    local response_code=$1
    local action=$2
    local module=$3
    local allow_404=${4:-false}  # Optional parameter to allow 404 responses

    if [ $response_code -eq 200 ] || [ $response_code -eq 204 ] || ([ "$allow_404" = "true" ] && [ $response_code -eq 404 ]); then
        return
    fi
    print_failure "$action" "$module" "$response_code"
}

# Delete a project in SonarCloud
delete() {
    MODULE=$1
    NAME=$(echo $MODULE | tr '_' '-')
    response=$(curl -s -w "%{http_code}" -X POST https://${SONAR_TOKEN}@sonarcloud.io/api/projects/delete \
        -d "name=testcontainers-go-${NAME}&project=testcontainers_testcontainers-go_${MODULE}&organization=testcontainers" 2>/dev/null)
    status_code=${response: -3}
    handle_curl_response $status_code "delete" "$MODULE"
    print_success "Deleted project" "$MODULE"
}

# Create all the projects in SonarCloud, starting with the modulegen module.
# Modules and examples need a prefix with the module type (module or example).
createAll() {
    create "modulegen"

    for MODULE in $MODULES; do
        create "modules_$MODULE"
    done

    for EXAMPLE in $EXAMPLES; do
        create "examples_$EXAMPLE"
    done
}

# create a new project in SonarCloud
create() {
    MODULE=$1
    NAME=$(echo $MODULE | tr '_' '-')
    response=$(curl -s -w "%{http_code}" -X POST https://${SONAR_TOKEN}@sonarcloud.io/api/projects/create \
        -d "name=testcontainers-go-${NAME}&project=testcontainers_testcontainers-go_${MODULE}&organization=testcontainers" 2>/dev/null)
    status_code=${response: -3}
    handle_curl_response $status_code "create" "$MODULE"
    print_success "Created project" "$MODULE"
}

# Rename all the main branches to the new name, starting with the modulegen module.
# Modules and examples need a prefix with the module type (module or example).
renameAllMainBranches() {
    rename_main_branch "modulegen"

    for MODULE in $MODULES; do
        rename_main_branch "modules_$MODULE"
    done

    for EXAMPLE in $EXAMPLES; do
        rename_main_branch "examples_$EXAMPLE"
    done
}

# rename the main branch to the new name: they originally have the name "master"
rename_main_branch() {
    MODULE=$1
    NAME=$(echo $MODULE | tr '_' '-')
    
    # Delete main branch (404 is acceptable here)
    response=$(curl -s -w "%{http_code}" -X POST https://${SONAR_TOKEN}@sonarcloud.io/api/project_branches/delete \
        -d "branch=main&project=testcontainers_testcontainers-go_${MODULE}&organization=testcontainers" 2>/dev/null)
    status_code=${response: -3}
    handle_curl_response $status_code "delete main branch" "$MODULE" true
    
    # Rename master to main
    response=$(curl -s -w "%{http_code}" -X POST https://${SONAR_TOKEN}@sonarcloud.io/api/project_branches/rename \
        -d "name=main&project=testcontainers_testcontainers-go_${MODULE}&organization=testcontainers" 2>/dev/null)
    status_code=${response: -3}
    handle_curl_response $status_code "rename branch to main" "$MODULE"
    print_success "Renamed branch to main" "$MODULE"
}

show_help() {
    echo "Usage: $0 [-h] [-a ACTION] [-p PROJECT_NAME]"
    echo
    echo "Options:"
    echo "  -h            Show this help message"
    echo "  -a ACTION     Action to perform (create, delete, createAll, deleteAll)"
    echo "  -p PROJECT    Project name to operate on (required for create/delete)"
    echo
    echo "Actions:"
    echo "  create        Creates a new SonarCloud project and sets up the main branch"
    echo "  delete        Deletes an existing SonarCloud project"
    echo "  createAll     Creates all projects and sets up their main branches"
    echo "  deleteAll     Deletes all existing projects"
    echo
    echo "Examples:"
    echo "  $0 -a create -p modules_mymodule    # Create a new project"
    echo "  $0 -a delete -p modules_mymodule    # Delete an existing project"
    echo "  $0 -a createAll                     # Create all projects"
    echo "  $0 -a deleteAll                     # Delete all projects"
}

validate_action() {
    local action=$1
    case $action in
        create|delete|createAll|deleteAll)
            return 0
            ;;
        *)
            echo "Error: Invalid action '$action'. Valid actions are: create, delete, createAll, deleteAll"
            return 1
            ;;
    esac
}

validate_project() {
    local action=$1
    local project=$2
    
    # Skip project validation for "All" actions
    if [[ "$action" == *"All" ]]; then
        return 0
    fi
    
    if [ -z "$project" ]; then
        echo "Error: Project name is required for action '$action'. Use -p to specify a project name"
        return 1
    fi
    return 0
}

main() {
    local project_name=""
    local action=""

    # Handle flags
    while getopts "ha:p:" opt; do
        case $opt in
            h)
                show_help
                exit 0
                ;;
            a)
                action="$OPTARG"
                if ! validate_action "$action"; then
                    exit 1
                fi
                ;;
            p)
                project_name="$OPTARG"
                ;;
            \?)
                echo "Invalid option: -$OPTARG" >&2
                show_help
                exit 1
                ;;
        esac
    done

    # Validate SONAR_TOKEN is set (except for help)
    if [ -z "${SONAR_TOKEN}" ]; then
        echo "Error: SONAR_TOKEN environment variable is not set"
        exit 1
    fi

    # Validate project name is provided for non-All actions
    if ! validate_project "$action" "$project_name"; then
        exit 1
    fi

    case $action in
        create)
            create "$project_name"
            rename_main_branch "$project_name"
            ;;
        delete)
            delete "$project_name"
            ;;
        createAll)
            createAll
            renameAllMainBranches
            ;;
        deleteAll)
            deleteAll
            ;;
        *)
            echo "No valid action specified. Use -h for help"
            ;;
    esac
}

main "$@"
