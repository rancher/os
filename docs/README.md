# Documentation for the `master` version of `rancher/os`

This dir is _not_ the documentation for any released version of RancherOS.
You can find that at the [Rancher Labs Documentation site](https://docs.rancher.com) and
specifically for [RancherOS](https://docs.rancher.com/os/).

When there are Pull Requests to the `rancher/os` repository that affect the user (or developer),
then it should include changes to the documenation in this directory.

When we make a new release of RancherOS, the `docs/os` dir will be copied into the `rancher/rancher.github.io`
repository to be accessible by users.

## Previewing your changes

You can either build and view your docs locally by running `make docs`, or you can
set your fork of the `rancher/os` repository to render your `master` using `GitHub Pages`.

To set up `GitHub Pages`, browse to your fork, then to the `Settings` - under `GitHub Pages`, set the `Source`
to `master branch /docs folder` and hit the `Save` button. GitHub will tell you the URL at which the
documenation can be read.
