#!/bin/bash
read -p "Title? " title
read -p "Hook? " hook
read -p "Tags? " -a tags
published_at=$(date -v+24H +"%Y-%m-%dT%H:%M:%S%z" | sed -E -n 's/([0-9]{2})([0-9]{2})$/\1:\2/p')

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
slug=$(slugify "$title")

# title="Hey what's up"
# slug="hey-whats-up"
# hook="Writing stuff"
# tags=("apple" "banana" "cherry")
# echo $title
# echo $slug
# echo $hook
# echo "${tags[0]}"  # Tag1
# echo "${tags[1]}"  # Tag2
# echo "${tags[2]}"  # Tag3
# echo $published_at

new_article_path=content/articles/$slug.md
image_dir_path=/content/images/$slug
mkdir -p content/images/$slug

cp scripts/new_article.md $new_article_path

# Now replace the values
sed -i '' "s/TITLE/$title/" $new_article_path
sed -i '' "s/HOOK/$hook/" $new_article_path
sed -i '' "s/PUBLISHED_AT/$published_at/" $new_article_path
sed -i '' "s|IMAGE|$image_dir_path|" "$new_article_path"
if [ ${#tags[@]} -eq 0 ]; then
  tags_str="[]"
else
  tags_str=$(printf '"%s", ' "${tags[@]}")
  tags_str="[${tags_str%, }]"
fi
sed -i '' "s/TAGS/$tags_str/" $new_article_path



