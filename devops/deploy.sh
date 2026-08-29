git config --local user.email "github-actions[bot]@users.noreply.github.com"
git config --local user.name "github-actions[bot]"

RELEASE_DATE=$(date +"%Y-%m-%d %H:%M:%S")

git tag -a "${{ env.NEW_TAG }}" -m "Release ${{ $NEW_RELEASE_TAG }} ($RELEASE_DATE)"