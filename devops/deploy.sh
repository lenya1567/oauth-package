git config --local user.email "github-actions[bot]@users.noreply.github.com"
git config --local user.name "github-actions[bot]"

RELEASE_DATE=$(date +"%Y-%m-%d %H:%M:%S")

git tag -a "$NEW_RELEASE_TAG" -m "Release $NEW_RELEASE_TAG ($RELEASE_DATE)"

echo "✅ Создан новый [$RELEASE_TYPE] релиз ($OLD_RELEASE_TAG -> $NEW_RELEASE_TAG)"

git push origin "$NEW_RELEASE_TAG"