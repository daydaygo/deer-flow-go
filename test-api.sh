#!/bin/bash

set -e

BASE_URL="${BASE_URL:-http://localhost:8001}"
FAILED=0
THREAD_ID=""

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

check_jq() {
    if ! command -v jq &> /dev/null; then
        echo "Warning: jq not found, JSON output will not be formatted"
        HAS_JQ=0
    else
        HAS_JQ=1
    fi
}

format_json() {
    if [ "$HAS_JQ" -eq 1 ]; then
        echo "$1" | jq '.' 2>/dev/null || echo "$1"
    else
        echo "$1"
    fi
}

test_endpoint() {
    local name="$1"
    local method="$2"
    local endpoint="$3"
    local data="$4"
    local expected_status="$5"
    
    echo -e "\nTesting: $name"
    echo "  $method $endpoint"
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "${BASE_URL}${endpoint}" 2>/dev/null)
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            "${BASE_URL}${endpoint}" 2>/dev/null)
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" = "$expected_status" ]; then
        echo -e "  ${GREEN}✓ PASSED${NC} (HTTP $http_code)"
        if [ -n "$body" ]; then
            echo "  Response:"
            format_json "$body" | sed 's/^/    /'
        fi
    else
        echo -e "  ${RED}✗ FAILED${NC} (Expected HTTP $expected_status, got $http_code)"
        if [ -n "$body" ]; then
            echo "  Response:"
            format_json "$body" | sed 's/^/    /'
        fi
        FAILED=1
    fi
    
    echo "$body"
}

test_1_health_check() {
    echo -e "\n========================================"
    echo "Test 1: Health Check"
    echo "========================================"
    
    test_endpoint "Health check" "GET" "/health" "" "200"
}

test_2_models_list() {
    echo -e "\n========================================"
    echo "Test 2: Models List"
    echo "========================================"
    
    test_endpoint "Models list" "GET" "/api/models" "" "200"
}

test_3_memory() {
    echo -e "\n========================================"
    echo "Test 3: Memory"
    echo "========================================"
    
    test_endpoint "Memory get" "GET" "/api/memory" "" "200"
}

test_4_thread_create() {
    echo -e "\n========================================"
    echo "Test 4: Thread Create"
    echo "========================================"
    
    local body
    body=$(test_endpoint "Thread create" "POST" "/api/langgraph/threads" '{"metadata": {"test": "true"}}' "200")
    
    if [ "$HAS_JQ" -eq 1 ] && [ -n "$body" ]; then
        THREAD_ID=$(echo "$body" | jq -r '.thread_id // .id // empty' 2>/dev/null)
        if [ -n "$THREAD_ID" ] && [ "$THREAD_ID" != "null" ]; then
            echo "  Created thread: $THREAD_ID"
        fi
    fi
}

test_5_thread_get() {
    echo -e "\n========================================"
    echo "Test 5: Thread Get"
    echo "========================================"
    
    if [ -z "$THREAD_ID" ]; then
        echo -e "  ${RED}✗ SKIPPED${NC} (No thread ID available)"
        return
    fi
    
    test_endpoint "Thread get" "GET" "/api/langgraph/threads/$THREAD_ID" "" "200"
}

test_6_run_create() {
    echo -e "\n========================================"
    echo "Test 6: Run Create"
    echo "========================================"
    
    if [ -z "$THREAD_ID" ]; then
        echo -e "  ${RED}✗ SKIPPED${NC} (No thread ID available)"
        return
    fi
    
    test_endpoint "Run create" "POST" "/api/langgraph/threads/$THREAD_ID/runs" \
        '{"assistant_id": "agent", "input": {"messages": [{"role": "user", "content": "hello"}]}}' "200"
}

test_7_run_stream() {
    echo -e "\n========================================"
    echo "Test 7: Run Stream"
    echo "========================================"
    
    if [ -z "$THREAD_ID" ]; then
        echo -e "  ${RED}✗ SKIPPED${NC} (No thread ID available)"
        return
    fi
    
    echo -e "Testing: Run stream"
    echo "  POST /api/langgraph/threads/$THREAD_ID/runs/stream"
    
    local http_code
    http_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
        -H "Content-Type: application/json" \
        -d '{"assistant_id": "agent", "input": {"messages": [{"role": "user", "content": "hello"}]}}' \
        "${BASE_URL}/api/langgraph/threads/$THREAD_ID/runs/stream" 2>/dev/null)
    
    if [ "$http_code" = "200" ] || [ "$http_code" = "202" ]; then
        echo -e "  ${GREEN}✓ PASSED${NC} (HTTP $http_code)"
    else
        echo -e "  ${RED}✗ FAILED${NC} (Expected HTTP 200/202, got $http_code)"
        FAILED=1
    fi
}

test_8_thread_delete() {
    echo -e "\n========================================"
    echo "Test 8: Thread Delete"
    echo "========================================"
    
    if [ -z "$THREAD_ID" ]; then
        echo -e "  ${RED}✗ SKIPPED${NC} (No thread ID available)"
        return
    fi
    
    test_endpoint "Thread delete" "DELETE" "/api/langgraph/threads/$THREAD_ID" "" "200"
}

print_summary() {
    echo -e "\n========================================"
    echo "Test Summary"
    echo "========================================"
    
    if [ "$FAILED" -eq 0 ]; then
        echo -e "${GREEN}All tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}Some tests failed.${NC}"
        exit 1
    fi
}

main() {
    echo "API Integration Tests"
    echo "Base URL: $BASE_URL"
    check_jq
    
    test_1_health_check
    test_2_models_list
    test_3_memory
    test_4_thread_create
    test_5_thread_get
    test_6_run_create
    test_7_run_stream
    test_8_thread_delete
    
    print_summary
}

main "$@"