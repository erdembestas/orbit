#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${ORBIT_API_BASE_URL:-http://localhost:8081}"
NAMESPACE="${NAMESPACE:-orbit-test}"

if ! command -v jq >/dev/null 2>&1; then
  echo "FAIL jq is required"
  exit 1
fi

FAILURES=0
TOKEN=""
GENERATED_NAMESPACE_PACK_ID=""
REQUEST_STATUS=""
REQUEST_FILE=""

pass() {
  echo "PASS $1"
}

skip() {
  echo "SKIP $1"
}

fail() {
  echo "FAIL $1"
  FAILURES=$((FAILURES + 1))
}

request() {
  local method="$1"
  local path="$2"
  local body="${3:-}"
  local auth="${4:-}"
  local output_file
  output_file="$(mktemp)"
  local headers=(-H "Accept: application/json")
  if [[ -n "$body" ]]; then
    headers+=(-H "Content-Type: application/json")
  fi
  if [[ -n "$auth" ]]; then
    headers+=(-H "Authorization: Bearer $auth")
  fi
  local status
  if [[ -n "$body" ]]; then
    status="$(curl -sS -o "$output_file" -w '%{http_code}' -X "$method" "${headers[@]}" -d "$body" "$BASE_URL$path")"
  else
    status="$(curl -sS -o "$output_file" -w '%{http_code}' -X "$method" "${headers[@]}" "$BASE_URL$path")"
  fi
  REQUEST_STATUS="$status"
  REQUEST_FILE="$output_file"
}

json_has() {
  local file="$1"
  local expr="$2"
  jq -e "$expr" "$file" >/dev/null 2>&1
}

run_required_check() {
  local label="$1"
  local method="$2"
  local path="$3"
  local expected_status="$4"
  local jq_expr="${5:-}"
  local body="${6:-}"
  local auth="${7:-}"
  request "$method" "$path" "$body" "$auth"
  local status="$REQUEST_STATUS"
  local file="$REQUEST_FILE"
  if [[ "$status" != "$expected_status" ]]; then
    fail "$label expected HTTP $expected_status got $status"
    rm -f "$file"
    return
  fi
  if [[ -n "$jq_expr" ]] && ! json_has "$file" "$jq_expr"; then
    fail "$label response validation failed"
    rm -f "$file"
    return
  fi
  pass "$label"
  rm -f "$file"
}

run_optional_get() {
  local label="$1"
  local path="$2"
  local auth="$3"
  request "GET" "$path" "" "$auth"
  local status="$REQUEST_STATUS"
  local file="$REQUEST_FILE"
  case "$status" in
    200) pass "$label" ;;
    404) skip "$label not implemented" ;;
    *) fail "$label expected HTTP 200 or 404 got $status" ;;
  esac
  rm -f "$file"
}

run_required_check "healthz" "GET" "/healthz" "200"
run_required_check "readyz" "GET" "/readyz" "200"
run_required_check "info" "GET" "/api/v1/info" "200" '.app_name == "orbit" and .cluster_name != null'

request "POST" "/api/v1/auth/login" '{"username":"admin","password":"admin"}'
if [[ "$REQUEST_STATUS" != "200" ]]; then
  fail "login admin/admin expected HTTP 200 got $REQUEST_STATUS"
else
  TOKEN="$(jq -r '.accessToken // empty' "$REQUEST_FILE")"
  if [[ -z "$TOKEN" || "$TOKEN" == "null" ]]; then
    fail "login admin/admin missing access token"
  else
    pass "login admin/admin"
  fi
fi
rm -f "$REQUEST_FILE"

if [[ -n "$TOKEN" ]]; then
  run_required_check "auth me" "GET" "/api/v1/auth/me" "200" '.username == "admin" and (.roles | index("admin") != null)' "" "$TOKEN"
  run_required_check "protected with token" "GET" "/api/v1/protected" "200" '.status == "ok"' "" "$TOKEN"
fi

run_required_check "protected without token" "GET" "/api/v1/protected" "401"
run_required_check "login admin/admin123 returns 401" "POST" "/api/v1/auth/login" "401" "" '{"username":"admin","password":"admin123"}'

if [[ -n "$TOKEN" ]]; then
  run_optional_get "clusters" "/api/v1/clusters" "$TOKEN"
  run_optional_get "finding rules" "/api/v1/finding-rules" "$TOKEN"
  run_optional_get "inventory resources" "/api/v1/inventory/resources" "$TOKEN"
  run_optional_get "findings" "/api/v1/findings" "$TOKEN"
  run_optional_get "evidence packs" "/api/v1/evidence-packs" "$TOKEN"
  run_optional_get "action plans" "/api/v1/action-plans" "$TOKEN"
  run_optional_get "controller status" "/api/v1/controller/status" "$TOKEN"
  run_required_check "cluster health" "GET" "/api/v1/cluster/health" "200" '.healthStatus != null and .healthScore != null and .metricsAvailable != null' "" "$TOKEN"
  run_required_check "cluster health nodes" "GET" "/api/v1/cluster/health/nodes" "200" 'type=="array"' "" "$TOKEN"
  run_required_check "cluster health namespaces" "GET" "/api/v1/cluster/health/namespaces" "200" 'type=="array"' "" "$TOKEN"
  run_required_check "cluster health history" "GET" "/api/v1/cluster/health/history?limit=10" "200" 'type=="array"' "" "$TOKEN"

  request "POST" "/api/v1/evidence-packs/generate" "{\"scopeType\":\"namespace\",\"namespace\":\"$NAMESPACE\",\"persist\":true}" "$TOKEN"
  case "$REQUEST_STATUS" in
    200)
      GENERATED_NAMESPACE_PACK_ID="$(jq -r '.id // empty' "$REQUEST_FILE")"
      if json_has "$REQUEST_FILE" '.token_estimate != null'; then
        pass "namespace evidence pack generation"
      else
        fail "namespace evidence pack missing token_estimate"
      fi
      ;;
    404) skip "namespace evidence pack generation not implemented" ;;
    *) fail "namespace evidence pack generation expected HTTP 200 or 404 got $REQUEST_STATUS" ;;
  esac
  rm -f "$REQUEST_FILE"

  if [[ -n "$GENERATED_NAMESPACE_PACK_ID" ]]; then
    run_required_check "get namespace evidence pack" "GET" "/api/v1/evidence-packs/$GENERATED_NAMESPACE_PACK_ID" "200" ".id == \"$GENERATED_NAMESPACE_PACK_ID\"" "" "$TOKEN"
    request "POST" "/api/v1/evidence-packs/$GENERATED_NAMESPACE_PACK_ID/reason" "" "$TOKEN"
    if [[ "$REQUEST_STATUS" == "200" ]]; then
      if json_has "$REQUEST_FILE" '.actionPlan.status == "draft" and .reasoning.rootCause != null'; then
        pass "reason over namespace evidence pack"
      else
        fail "reason over namespace evidence pack response validation failed"
      fi
    else
      fail "reason over namespace evidence pack expected HTTP 200 got $REQUEST_STATUS"
    fi
    rm -f "$REQUEST_FILE"
    run_optional_get "action plans after reasoning" "/api/v1/action-plans" "$TOKEN"
  fi

  if command -v kubectl >/dev/null 2>&1; then
    POD="$(kubectl -n "$NAMESPACE" get pod -l app.kubernetes.io/name=orbit-api -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)"
    if [[ -n "$POD" ]]; then
      request "POST" "/api/v1/evidence-packs/generate" "{\"scopeType\":\"pod\",\"namespace\":\"$NAMESPACE\",\"name\":\"$POD\",\"persist\":true}" "$TOKEN"
      case "$REQUEST_STATUS" in
        200)
          if json_has "$REQUEST_FILE" '.token_estimate != null' && ! grep -Eiq 'secret|password|jwt' "$REQUEST_FILE"; then
            pass "pod evidence pack generation"
          else
            fail "pod evidence pack validation failed"
          fi
          ;;
        404) skip "pod evidence pack generation not implemented" ;;
        *) fail "pod evidence pack generation expected HTTP 200 or 404 got $REQUEST_STATUS" ;;
      esac
      rm -f "$REQUEST_FILE"
    else
      skip "pod evidence pack generation skipped because orbit-api pod was not found"
    fi
  fi
fi

if [[ "$FAILURES" -ne 0 ]]; then
  exit 1
fi
