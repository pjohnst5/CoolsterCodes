#!/bin/bash
read -p "Article title? " article

slugify() {
  local input="$1"

  input=$(echo "$input" \
    | tr '[:upper:]' '[:lower:]' \
    | sed 's/[^a-z0-9[:space:]-]//g' \
    | sed 's/^[[:space:]]*//;s/[[:space:]]*$//' \
    | sed -E 's/[[:space:]]+/-/g' \
    | sed -E 's/-+/-/g')

  echo "$input"
}

slug=$(slugify "$article")


