#!/usr/bin/env bash
#
# terraform plan の結果を PR にコメントする。
#
# usage: scripts/comment_plan.sh <env> <plan_file> <pr_number>

set -euo pipefail

env="${1:?usage: $0 <env> <plan_file> <pr_number>}"
plan_file="${2:?usage: $0 <env> <plan_file> <pr_number>}"
pr_number="${3:?usage: $0 <env> <plan_file> <pr_number>}"

# コメントの上限は 65536 文字
max_length=60000
plan="$(cat "${plan_file}")"
if [ "${#plan}" -gt "${max_length}" ]; then
  plan="${plan:0:${max_length}}
...(省略。全文は Actions のログを参照)"
fi

body="### Terraform Plan (\`${env}\`)

<details><summary>結果を表示</summary>

\`\`\`hcl
${plan}
\`\`\`

</details>"

gh pr comment "${pr_number}" --body "${body}"
