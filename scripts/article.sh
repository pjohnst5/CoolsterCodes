#!/bin/bash
read -p "Slug? " slug
read -p "Title? " title
read -p "Hook? " hook
read -p "Tags? " -a tags

published_at=$(date -v+8H +"%Y-%m-%dT%H:%M:%S%z" | sed -E -n 's/([0-9]{2})([0-9]{2})$/\1:\2/p')
new_article_dir=content/articles/$slug
new_article_path=$new_article_dir/$slug.md
mkdir -p $new_article_dir

cp scripts/new_article.md $new_article_path

# Now replace the values
sed -i '' "s/TITLE/$title/" $new_article_path
sed -i '' "s/HOOK/$hook/" $new_article_path
sed -i '' "s/PUBLISHED_AT/$published_at/" $new_article_path
if [ ${#tags[@]} -eq 0 ]; then
  tags_str="[]"
else
  tags_str=$(printf '"%s", ' "${tags[@]}")
  tags_str="[${tags_str%, }]"
fi
sed -i '' "s/TAGS/$tags_str/" $new_article_path



