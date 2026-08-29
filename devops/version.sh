git fetch --tags

LATEST_TAG=$(git tag -l "v*" | sort -V | tail -n1)

if [ -z "$LATEST_TAG" ]; then
    LATEST_TAG="v0.0.0"
fi

echo "Release type: $RELEASE_TYPE"
echo "Old version: $LATEST_TAG"

VERSION_NUMBERS=${LATEST_TAG#v}
IFS='.' read -r major minor patch <<< "$VERSION_NUMBERS"

NEW_PATCH="$patch"
NEW_MINOR="$minor"
NEW_MAJOR="$major"

case "$RELEASE_TYPE" in
    "patch")
        NEW_PATCH=$((patch + 1))
        ;;
    "minor")
        NEW_MINOR=$((minor + 1))
        ;;
    "major")
        NEW_MAJOR=$((major + 1))
        ;;
    *)
        echo "Unknown release type"
        exit 1
        ;;
esac

NEW_VERSION="v$NEW_MAJOR.$NEW_MINOR.$NEW_PATCH"

echo "New version: $NEW_VERSION"
echo "NEW_RELEASE_TAG=$NEW_VERSION" >> $GITHUB_ENV