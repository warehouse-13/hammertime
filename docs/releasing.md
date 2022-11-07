# Releasing hammertime

## Determine release version

The projects follows [semantic versioning](https://semver.org/#semantic-versioning-200)
(sort of).
The next appropriate version number will depend on what is going into the release.

First pull all tags:

```bash
git pull --tags

git describe --tags --abbrev=0
```

This will give you the latest release, the next release will increment from here.

## Create tag

* Checkout upstream main
* Create a tag with the version number:

```bash
# assuming the answer to git describe was v0.0.9
RELEASE_VERSION=v0.0.10
git tag -s "${RELEASE_VERSION}" -m "${RELEASE_VERSION}"
```

* Push the tag (to upstream if working from a fork)

``` bash
git push origin "${RELEASE_VERSION}"
```

* Check the [release](https://github.com/warehouse-13/hammertime/actions/workflows/release.yml)
  GitHub Actions workflow completes successfully.

## Edit & Publish GitHub Release

* Go to the draft release in GitHub.
* Make any edits to generated release notes
  * If there are any breaking changes then manually add a note at the beginning
    of the release notes informing the user what they need to be aware of/do.
  * Sometimes you may want to combine changes into 1 line
