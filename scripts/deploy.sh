#!/usr/bin/env bash
#
# Lambda のコードを更新する。
#
# 新バージョンを発行してから疎通確認し、通った場合だけエイリアスを切り替える。
# 確認に失敗した場合はエイリアスが旧バージョンを指したままなので、利用者には影響しない。
#
# usage: scripts/deploy.sh <env>

set -euo pipefail

env="${1:?usage: $0 <env>}"
tf_dir="terraform/env/${env}"

function_name="$(terraform -chdir="${tf_dir}" output -raw function_name)"
alias_name="$(terraform -chdir="${tf_dir}" output -raw alias_name)"

echo "--- publish version"

version="$(aws lambda update-function-code \
  --function-name "${function_name}" \
  --zip-file fileb://fn.zip \
  --publish \
  --query Version --output text --no-cli-pager)"
aws lambda wait function-updated --function-name "${function_name}:${version}"

echo "published: ${function_name}:${version}"
echo "--- smoke test"

# エイリアス経由ではなくバージョンを直接呼ぶため、公開中の挙動には影響しない
health_event='{
  "version": "2.0",
  "rawPath": "/health",
  "requestContext": { "http": { "method": "GET", "path": "/health" } }
}'
aws lambda invoke \
  --function-name "${function_name}:${version}" \
  --payload "${health_event}" \
  --cli-binary-format raw-in-base64-out \
  --no-cli-pager \
  response.json > /dev/null

cat response.json

grep -q '"statusCode":200' response.json

echo "--- switch alias"

aws lambda update-alias \
  --function-name "${function_name}" \
  --name "${alias_name}" \
  --function-version "${version}" \
  --no-cli-pager > /dev/null

echo "switched: ${alias_name} -> ${version}"
