#!/usr/bin/env bash
# Medhen Phase 0 — Motor buy→claim storyboard (EIC)
# Supports monolith or mesh; optional Keycloak JWT when KEYCLOAK_URL is set.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BASE="${MEDHEN_URL:-http://localhost:8080/api/v1}"
SEED="${ROOT}/seeds/demo-personas.json"

AUTH=()
if [[ -n "${KEYCLOAK_URL:-}" ]]; then
  TOKEN=$(KEYCLOAK_USER="${KEYCLOAK_USER:-demo-agent}" KEYCLOAK_PASS="${KEYCLOAK_PASS:-medhen-demo}" \
    "${ROOT}/scripts/keycloak-token.sh")
  AUTH=(-H "Authorization: Bearer ${TOKEN}")
  echo "== Auth: Keycloak JWT (demo-agent) =="
else
  AUTH=(-H "X-User-ID: demo-agent")
  echo "== Auth: demo headers =="
fi

HDR=(-H "Content-Type: application/json" -H "X-Tenant-ID: eic" -H "Accept-Language: am" "${AUTH[@]}")

read_seed() {
  python3 -c "
import json, sys
d=json.load(open('$SEED'))
p=next(x for x in d['personas'] if x['id']=='abebe')
v=p['vehicle']
f=d['fnol']
print(json.dumps({'party':p,'vehicle':v,'fnol':f}))
"
}

PAYLOAD=$(read_seed)
PARTY_JSON=$(python3 -c "import json,sys; d=json.load(sys.stdin); p=d['party']; print(json.dumps({'fullName':p['fullName'],'fullNameAm':p['fullNameAm'],'phoneE164':p['phoneE164'],'email':p['email'],'address':p['address']}))" <<<"$PAYLOAD")

echo "== Health =="
curl -sf "${MEDHEN_URL:-http://localhost:8080}/health" | tee /tmp/medhen-health.json
echo

echo "== 1. Register party =="
PARTY=$(curl -sf "${BASE}/parties" "${HDR[@]}" \
  -H "Idempotency-Key: $(uuidgen)" \
  -d "$PARTY_JSON")
echo "$PARTY" | tee /tmp/medhen-party.json
PARTY_ID=$(python3 -c "import sys,json; print(json.load(sys.stdin)['id'])" <<<"$PARTY")

echo "== 2. Product + risk schema =="
curl -sf "${BASE}/products/MOTOR-PRIVATE-COMP" "${HDR[@]}" | tee /tmp/medhen-product.json
echo
curl -sf "${BASE}/products/MOTOR-PRIVATE-COMP/risk-schema" "${HDR[@]}" | tee /tmp/medhen-schema.json
echo

RISK=$(python3 -c "import json,sys; v=json.load(sys.stdin)['vehicle']; v['sumInsuredMinor']=int(v.pop('sumInsuredETB')*100); print(json.dumps(v))" <<<"$PAYLOAD")

echo "== 3. Create quote (STP) =="
QUOTE=$(curl -sf "${BASE}/quotes" "${HDR[@]}" \
  -H "Idempotency-Key: $(uuidgen)" \
  -d "{\"partyId\":\"${PARTY_ID}\",\"productCode\":\"MOTOR-PRIVATE-COMP\",\"risk\":${RISK}}")
echo "$QUOTE" | tee /tmp/medhen-quote.json
QUOTE_ID=$(python3 -c "import sys,json; print(json.load(sys.stdin)['id'])" <<<"$QUOTE")
TOTAL=$(python3 -c "import sys,json; print(json.load(sys.stdin)['totalMinor'])" <<<"$QUOTE")
echo "Premium total (minor): $TOTAL"

echo "== 4. Bind quote =="
BIND=$(curl -sf "${BASE}/quotes/${QUOTE_ID}/bind" "${HDR[@]}" -X POST -H "Idempotency-Key: $(uuidgen)" -d '{}')
echo "$BIND" | tee /tmp/medhen-bind.json
INVOICE_ID=$(python3 -c "import sys,json; print(json.load(sys.stdin)['invoice']['id'])" <<<"$BIND")
POLICY_ID=$(python3 -c "import sys,json; print(json.load(sys.stdin)['policy']['id'])" <<<"$BIND")

echo "== 5. Pay via Telebirr =="
PHONE=$(python3 -c "import json,sys; print(json.load(sys.stdin)['party']['phoneE164'])" <<<"$PAYLOAD")
PAY=$(curl -s "${BASE}/billing/invoices/${INVOICE_ID}/pay" "${HDR[@]}" \
  -H "Idempotency-Key: $(uuidgen)" \
  -d "{\"channel\":\"telebirr\",\"phone\":\"${PHONE}\"}")
echo "$PAY" | tee /tmp/medhen-pay.json
POLICY_NO=$(python3 -c "import sys,json; print(json.load(sys.stdin)['policy']['policyNumber'])" <<<"$PAY")
echo "Issued policy: $POLICY_NO"

echo "== 6. FNOL fast-track =="
FNOL_BODY=$(python3 -c "
import json,sys,datetime
d=json.load(sys.stdin)
f=d['fnol']
print(json.dumps({
  'policyId':'${POLICY_ID}',
  'lossDate': datetime.datetime.utcnow().strftime('%Y-%m-%dT%H:%M:%SZ'),
  'description': f['description'],
  'latitude': f['latitude'],
  'longitude': f['longitude'],
  'estimatedAmountMinor': int(f['estimatedAmountETB']*100),
  'photoObjectKeys': ['fnol/demo-photo-1.jpg']
}))
" <<<"$PAYLOAD")
CLAIM=$(curl -sf "${BASE}/claims" "${HDR[@]}" \
  -H "Idempotency-Key: $(uuidgen)" \
  -d "$FNOL_BODY")
echo "$CLAIM" | tee /tmp/medhen-claim.json
CLAIM_ID=$(python3 -c "import sys,json; print(json.load(sys.stdin)['id'])" <<<"$CLAIM")

echo "== 7. Settle =="
curl -sf "${BASE}/claims/${CLAIM_ID}/settle" "${HDR[@]}" -X POST -H "Idempotency-Key: $(uuidgen)" -d '{}' | tee /tmp/medhen-settle.json
echo

echo "== 8. Audit + KPIs =="
curl -sf "${BASE}/audit?limit=20" "${HDR[@]}" | tee /tmp/medhen-audit.json
echo
curl -sf "${BASE}/demo/kpis" "${HDR[@]}" | tee /tmp/medhen-kpis.json
echo

# Postgres verification when DATABASE_URL is set on the runner
if [[ -n "${DATABASE_URL:-}" ]]; then
  echo "== Postgres: policies issued =="
  psql "$DATABASE_URL" -t -c "SELECT count(*) FROM ${PG_SCHEMA:-pc_medhen}.policies WHERE status='ISSUED';" 2>/dev/null || \
    echo "(psql not available — skip DB check)"
fi

echo "OK — storyboard complete. Policy $POLICY_NO"
