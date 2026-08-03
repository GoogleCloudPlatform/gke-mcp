#!/usr/bin/env bash
# preflight.sh — puerta pre-deploy. Corre DENTRO de un repo git.
# Falla (exit 1) si hay secretos/.env trackeados o falta higiene básica.
set -uo pipefail
cd "$(git rev-parse --show-toplevel 2>/dev/null || echo .)"
FAIL=0; WARN=0
red(){ printf '\033[31m✗ %s\033[0m\n' "$1"; FAIL=1; }
yel(){ printf '\033[33m▲ %s\033[0m\n' "$1"; WARN=1; }
grn(){ printf '\033[32m✓ %s\033[0m\n' "$1"; }
tracked=$(git ls-files)
envs=$(echo "$tracked" | grep -iE '(^|/)\.env($|\.)' | grep -viE '\.(example|template|sample)$' || true)
[ -n "$envs" ] && red ".env trackeado(s):
$(echo "$envs" | sed 's/^/    /')" || grn "sin .env trackeados"
echo "$tracked" | grep -qiE '(^|/)\.gitignore$' && grn ".gitignore presente" || red "falta .gitignore"
keys=$(echo "$tracked" | grep -iE '\.pem$|\.key$|(^|/)id_rsa|\.p12$|\.pfx$|\.pkcs12$|\.keystore$|(^|/)\.pypirc$' | grep -viE '\.(example|template|sample)$' || true)
[ -n "$keys" ] && red "posible material de clave trackeado:
$(echo "$keys" | sed 's/^/    /')" || grn "sin ficheros de clave trackeados"
for f in $(echo "$tracked" | grep -E '(^|/)\.npmrc$' || true); do
  git show "HEAD:$f" 2>/dev/null | grep -qiE '_authToken|_password|:_auth' && red ".npmrc con credenciales: $f" || grn ".npmrc sin token: $f"
done
PAT='sk_live_[0-9A-Za-z]{16,}|sk_test_[0-9A-Za-z]{16,}|AKIA[0-9A-Z]{16}|ghp_[A-Za-z0-9]{36}|gho_[A-Za-z0-9]{36}|AIza[0-9A-Za-z_-]{35}|xox[baprs]-[0-9A-Za-z-]{10,}|-----BEGIN (RSA |OPENSSH |EC )?PRIVATE KEY-----|[0-9]{8,10}:[A-Za-z0-9_-]{35}'
hits=$(echo "$tracked" | grep -iE '\.(js|ts|jsx|tsx|mjs|cjs|py|rs|go|rb|php|java|json|ya?ml|toml|env|sh|txt|md|astro|vue|svelte)$' \
  | grep -viE 'package-lock|pnpm-lock|yarn\.lock|node_modules' \
  | while IFS= read -r f; do git show "HEAD:$f" 2>/dev/null | grep -EnI "$PAT" | sed "s|^|$f:|"; done)
[ -n "$hits" ] && red "posible secreto hardcodeado:
$(echo "$hits" | sed -E 's/(.{60}).*/\1…/' | sed 's/^/    /')" || grn "sin secretos hardcodeados evidentes"
arte=$(echo "$tracked" | grep -E '(^|/)(node_modules|dist|\.vercel|\.next|\.output)/' | head -3 || true)
[ -n "$arte" ] && yel "artefactos de build trackeados (ej.):
$(echo "$arte" | sed 's/^/    /')"
echo "------------------------------------------------------------"
if [ "$FAIL" -ne 0 ]; then echo "PREFLIGHT: FAIL — no desplegar"; exit 1
elif [ "$WARN" -ne 0 ]; then echo "PREFLIGHT: PASS con avisos"; exit 0
else echo "PREFLIGHT: PASS"; exit 0; fi
