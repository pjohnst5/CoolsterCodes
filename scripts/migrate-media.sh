#!/usr/bin/env bash
#
# One-time migration:
#   1. Verify prerequisites (az login, container exists, public read access).
#   2. Upload all currently-committed media to Azure Blob Storage via `make sync-media`.
#   3. `git rm --cached` all migrated media so it stops being tracked by git
#      (local copies remain on disk; they're now gitignored).
#
# Re-runnable: the sync step is idempotent (skips unchanged blobs), and
# git rm --cached is a no-op for already-untracked files.

set -euo pipefail

SUBSCRIPTION="65ce76d9-2f1b-4c0f-849c-163ddd03b7a9"
RESOURCE_GROUP="CoolsterCodes"
STORAGE_ACCOUNT="coolstercodes"
CONTAINER="public"

echo "==> Checking Azure CLI login..."
if ! az account show >/dev/null 2>&1; then
  echo "❌ Not logged in. Run: az login"
  exit 1
fi
az account set --subscription "$SUBSCRIPTION"

echo "==> Enabling public blob access on storage account (idempotent)..."
az storage account update \
  --resource-group "$RESOURCE_GROUP" \
  --name "$STORAGE_ACCOUNT" \
  --allow-blob-public-access true \
  --output none

echo "==> Setting container '$CONTAINER' to public blob read (idempotent)..."
az storage container set-permission \
  --name "$CONTAINER" \
  --account-name "$STORAGE_ACCOUNT" \
  --auth-mode login \
  --public-access blob \
  --output none

echo "==> Granting 'Storage Blob Data Contributor' to signed-in user (idempotent)..."
PRINCIPAL_ID=$(az ad signed-in-user show --query id -o tsv)
SCOPE="/subscriptions/${SUBSCRIPTION}/resourceGroups/${RESOURCE_GROUP}/providers/Microsoft.Storage/storageAccounts/${STORAGE_ACCOUNT}"
az role assignment create \
  --role "Storage Blob Data Contributor" \
  --assignee "$PRINCIPAL_ID" \
  --scope "$SCOPE" \
  --output none 2>/dev/null || echo "   (role already assigned or requires manual grant)"

echo "==> Installing coolstercodes CLI..."
make install

echo "==> Uploading all local media to the '$CONTAINER' container..."
make sync-media

echo "==> git rm --cached tracked media files..."
patterns=(
  "content/articles/**/*.png"   "content/articles/**/*.jpg"  "content/articles/**/*.jpeg"
  "content/articles/**/*.gif"   "content/articles/**/*.svg"  "content/articles/**/*.webp"
  "content/articles/**/*.pdf"   "content/articles/**/*.mp4"  "content/articles/**/*.mov"
  "content/articles/**/*.webm"  "content/articles/**/*.mp3"  "content/articles/**/*.wav"
  "content/articles/**/*.doc"   "content/articles/**/*.docx" "content/articles/**/*.xls"
  "content/articles/**/*.xlsx"  "content/articles/**/*.ppt"  "content/articles/**/*.pptx"
  "content/articles/**/*.zip"
  "content/pages/**/*.png"      "content/pages/**/*.jpg"     "content/pages/**/*.jpeg"
  "content/pages/**/*.gif"      "content/pages/**/*.svg"     "content/pages/**/*.webp"
  "content/pages/**/*.pdf"      "content/pages/**/*.mp4"     "content/pages/**/*.mov"
  "content/pages/**/*.webm"     "content/pages/**/*.mp3"     "content/pages/**/*.wav"
  "content/pages/**/*.doc"      "content/pages/**/*.docx"    "content/pages/**/*.xls"
  "content/pages/**/*.xlsx"     "content/pages/**/*.ppt"     "content/pages/**/*.pptx"
  "content/pages/**/*.zip"
  "content/images/*.png"        "content/images/*.jpg"       "content/images/*.jpeg"
  "content/images/*.gif"        "content/images/*.svg"       "content/images/*.webp"
  "content/images/*.ico"
)

removed=0
for pat in "${patterns[@]}"; do
  # `git ls-files` expands the pattern relative to the repo root.
  while IFS= read -r f; do
    [ -z "$f" ] && continue
    git rm --cached --quiet -- "$f"
    removed=$((removed + 1))
  done < <(git ls-files -- "$pat" 2>/dev/null || true)
done

echo "==> Removed $removed media files from git tracking (local files preserved)."
echo
echo "✅ Migration complete."
echo
echo "Next steps:"
echo "  1. Verify a blob loads, e.g.:"
echo "     https://${STORAGE_ACCOUNT}.blob.core.windows.net/${CONTAINER}/content/images/favicon.png"
echo "  2. Run 'make build' and spot-check a rendered article."
echo "  3. Commit the .gitignore, code changes, and the deletions with a message like:"
echo "     'Move media to Azure Blob Storage'"
