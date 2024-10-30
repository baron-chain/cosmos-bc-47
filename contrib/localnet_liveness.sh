#!/usr/bin/env bash

set -euo pipefail

# Constants
readonly CURL_TIMEOUT=10
readonly CHAOS_INTERVAL=10
readonly EXIT_SUCCESS=0
readonly EXIT_FAILURE=1

# Functions
usage() {
    cat <<EOF
Usage: $0 <iterations> <sleep_seconds> <target_blocks> <node_address>

Arguments:
    iterations     - Number of iterations to run
    sleep_seconds  - Seconds to sleep between iterations
    target_blocks  - Block height to declare completion
    node_address   - Node address to poll (without port)

Example: $0 100 5 1000 http://localhost
EOF
    exit ${EXIT_FAILURE}
}

validate_inputs() {
    local -r iterations=$1 sleep_seconds=$2 target_blocks=$3 node_address=$4
    
    [[ $iterations =~ ^[0-9]+$ ]] || { echo "Error: iterations must be a number"; exit ${EXIT_FAILURE}; }
    [[ $sleep_seconds =~ ^[0-9]+$ ]] || { echo "Error: sleep seconds must be a number"; exit ${EXIT_FAILURE}; }
    [[ $target_blocks =~ ^[0-9]+$ ]] || { echo "Error: target blocks must be a number"; exit ${EXIT_FAILURE}; }
    [[ $node_address =~ ^https?:// ]] || { echo "Error: node address must start with http:// or https://"; exit ${EXIT_FAILURE}; }
}

get_block_height() {
    local -r node_address=$1
    curl -s -m ${CURL_TIMEOUT} "${node_address}:26657/status" | jq -r '.result.sync_info.latest_block_height' || echo ""
}

restart_random_container() {
    local -r containers=("$@")
    local -r container_count=${#containers[@]}
    
    if ((container_count > 0)); then
        local -r random_index=$((RANDOM % container_count))
        local -r container=${containers[random_index]}
        echo "Restarting random docker container: ${container}"
        docker restart "${container}" &>/dev/null &
    fi
}

main() {
    # Validate command line arguments
    [[ $# -eq 4 ]] || usage
    
    local -r iterations=$1
    local -r sleep_seconds=$2
    local -r target_blocks=$3
    local -r node_address=$4
    
    validate_inputs "$@"
    
    # Get docker containers
    mapfile -t docker_containers < <(docker ps -q -f name=simd --format='{{.Names}}')
    
    # Main loop
    local count=0
    while ((count < iterations)); do
        local current_block
        current_block=$(get_block_height "${node_address}")
        
        if [[ -n ${current_block} ]]; then
            echo "Current block height: ${current_block}"
            
            if ((current_block > target_blocks)); then
                echo "Target block height reached. Success!"
                exit ${EXIT_SUCCESS}
            fi
        fi
        
        # Network chaos: restart random container every CHAOS_INTERVAL iterations
        if ((count % CHAOS_INTERVAL == 0)); then
            restart_random_container "${docker_containers[@]}"
        fi
        
        ((count++))
        sleep "${sleep_seconds}"
    done
    
    echo "Timeout reached after ${iterations} iterations. Failure!"
    exit ${EXIT_FAILURE}
}

main "$@"
