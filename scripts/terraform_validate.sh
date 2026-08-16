#!/usr/bin/env bash
#
# 全環境の構文チェックを行う (CI 専用)
# .terraform を消すため、ローカルでは実行しない
#
# usage: scripts/terraform_validate.sh

set -euo pipefail

for dir in terraform/env/* terraform/initial_setup/env/*; do
  echo "::group::${dir}"
  rm -rf "${dir}/.terraform"
  terraform -chdir="${dir}" init -backend=false -input=false
  terraform -chdir="${dir}" validate
  echo "::endgroup::"
done
