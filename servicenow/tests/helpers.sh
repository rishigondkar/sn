# Test helpers for API Gateway BFF E2E tests.
# Source from run.sh:  . "$(dirname "$0")/helpers.sh"

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
USER_ID="${X_USER_ID:-e2e-test-user}"
USER_NAME="${X_USER_NAME:-E2E Test User}"

# Curl with common headers (JSON, auth). Usage: curl_api METHOD PATH [BODY]
curl_api() {
  local method="$1"
  local path="$2"
  local body="${3:-}"
  if [[ -n "$body" ]]; then
    curl -s -w "\n%{http_code}" -X "$method" \
      -H "X-User-Id: $USER_ID" \
      -H "X-User-Name: $USER_NAME" \
      -H "Content-Type: application/json" \
      -d "$body" \
      "$BASE_URL$path"
  else
    curl -s -w "\n%{http_code}" -X "$method" \
      -H "X-User-Id: $USER_ID" \
      -H "X-User-Name: $USER_NAME" \
      "$BASE_URL$path"
  fi
}

# Split response: last line = status, rest = body
# Usage: response=$(curl_api GET /health); status=$(echo "$response" | tail -n1); body=$(echo "$response" | sed '$d')
get_status() { echo "$1" | tail -n1; }
get_body()   { echo "$1" | sed '$d'; }

# Assert HTTP status. Usage: assert_status "$full_response" 200
assert_status() {
  local response="$1"
  local want="$2"
  local status; status=$(get_status "$response")
  if [[ "$status" != "$want" ]]; then
    echo "FAIL expected HTTP $want, got $status"
    return 1
  fi
  return 0
}

# Assert body contains string. Usage: assert_contains "$body" '"id"'
assert_contains() {
  local body="$1"
  local substr="$2"
  if [[ "$body" != *"$substr"* ]]; then
    echo "FAIL body does not contain: $substr"
    return 1
  fi
  return 0
}

# Assert body contains error code. Usage: assert_error_code "$body" "VALIDATION_ERROR"
assert_error_code() {
  local body="$1"
  local code="$2"
  assert_contains "$body" "$code"
}

# Run one test: run_test "Test name" command_that_returns_0_on_success
run_test() {
  local name="$1"
  shift
  local out
  if out=$("$@" 2>&1); then
    echo "PASS $name"
    return 0
  else
    echo "FAIL $name"
    echo "$out" | sed 's/^/  /'
    return 1
  fi
}

# Counters for summary (call from run.sh)
PASSED=0
FAILED=0

record_pass() { PASSED=$((PASSED + 1)); }
record_fail() { FAILED=$((FAILED + 1)); }

print_summary() {
  local total=$((PASSED + FAILED))
  echo ""
  echo "--- Summary ---"
  echo "Passed: $PASSED  Failed: $FAILED  Total: $total"
  if [[ $FAILED -gt 0 ]]; then
    return 1
  fi
  return 0
}
