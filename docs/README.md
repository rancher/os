# Rancher Labs Documentation

This README file includes information pertaining to the documentation project for both Rancher and Rancher OS.

## Rancher

Rancher is an open source project that provides a complete platform for operating Docker in production. It provides infrastructure services such as multi-host networking, global and local load balancing, and volume snapshots. It integrates native Docker management capabilities such as Docker Machine and Docker Swarm. It offers a rich user experience that enables devops admins to operate Docker in production at large scale.

## Rancher OS

Operating system composed entirely of Docker containers. Everything in RancherOS is a container managed by Docker. This includes system services such as udev and syslog. RancherOS includes only the bare minimum amount of software needed to run Docker.

## Rancher Labs Documentation website

Rancher documentation is available at <http://rancher.com/docs/rancher/>.

As Rancher has gone GA, we've added in version control. The default Rancher docs site will always be referring to the latest release. We will branch off specific versions of Rancher that are deemed GA, which would also be tagged as `rancher/server:stable`.

Currently, we've added support for Chinese version of the docs site per community request. Currently, Rancher will not be actively translating the docs site, but we welcome PRs.

Rancher OS documentation is available at <http://rancher.com/docs/os/>.

## Contributing to Rancher Labs Documentation Project

### About Rancher Labs Documentation Site

Rancher Labs documentation is hosted on GitHub Pages and published online by using Jekyll, an easy blog-aware static website generator. For more details on how to set up Jekyll, we recommend you to read <https://help.github.com/articles/using-jekyll-with-pages/>. If you are using Windows, we strongly advise you to follow the instruction given at <http://jekyllrb.com/docs/windows/>.

For information on editing `.md` files (Markdown), refer to <https://daringfireball.net/projects/markdown/syntax>.

Or you can use the `make live` Makefile target (or run `docker run --rm -it -p 80:4000 -v $(PWD):/build rancher/rancher.github.io:build jekyll serve -w -P 4000 --incremental` by hand to use the Jekyll build image used for our production pipeline.

### Setting up the Git Environment

In your browser, navigate to <https://github.com/rancher/rancher.github.io>.

Fork this repository by clicking the Fork button on the upper right-hand side. A fork of the repository is generated for you. On the right-hand side of the page of your fork, under 'HTTPS clone URL', copy the URL by clicking the Copy to clipboard icon.

On your computer, follow these steps to setup a local repository to start working on the documentation:

```shell
git clone https://github.com/YOUR_ACCOUNT/rancher.github.io.git
cd rancher.github.io
git remote add upstream https://github.com/rancher/rancher.github.io
git checkout master
git fetch upstream
git merge upstream/master
```

### Updating the Files

We recommend you to create a new branch to update the documentation files and that you do not disturb the master branch, other than pulling in changes from `upstream/master`.
For example, you create a branch, `dev`, on your computer to make changes locally to the documentation. This `dev` branch will be your local repository which then be pushed to your forked repository on GitHub where you will create a Pull Request for the changes to be committed into the official documentation.

It is a healthy practice to create a new branch each time you want to contribute to the documentation and only track the changes for that pull request in this branch.

```shell
git checkout -b dev
```

The argument `-b dev` creates a new branch named `dev`. Now you can make necessary changes to the documentation.

```shell
git add .
git commit -a -m "commit message for your changes"
```

You can optionally run Jekyll locally on your computer to be able to see the final result of your modifications and you write them. For that, use the command below. You can refer to [Jekyll's official website](https://jekyllrb.com/) for more details.

```shell
jekyll serve
```

Additionally, you can use the provided `Makefile` to build and test in a Docker container:

```shell
make live
```

### Merging upstream/master into Your Local Branch (`dev`)

Maintain an up-to-date master branch in your local repository. Merge the changes on a daily basis from the `upstream/master` (the official documentation repository) into your local repository.

Ensure that you do complete this activity before you start working on a feature as well as right before you submit your changes as a pull request.

You can also do this process periodically while you work on your changes to make sure that you are working off the most recent version of the documentation.

```shell
# Checkout your local 'master' branch.
git checkout master

# Synchronize your local repository with 'upstream/master', so that you have all the latest changes.
git fetch upstream

# Merge the latest changes from the 'upstream/master' into your local 'master' branch to make it up-to-date.
git merge upstream/master

# Checkout your local development branch (e.g.: 'dev').
git checkout dev

# Pull the latest changes into your local development branch.
git pull master
```

### Checklist for contributions

Please check the following list before you submit a Pull Request to make sure we can approve it right away!

* Check if your change applies to more than one version, and if so, please make the same change in other versions as well. (Rancher only)
* If your change only applies to a minor version, make sure it is specified in the text, i.e. `Available as of Rancher v1.6.6` or `Available as of RancherOS v1.1.0`.
* If your change is regarding an item in the `rancher-catalog`, make sure the change is also applied in the README there.

### Making a Pull Request on GitHub

**Important:** Ensure that you have merged `upstream/master` into your dev branch before you do the following.

After you have made necessary changes to the documentation and are ready to contribute them, create a Pull Request on GitHub. You do it by pushing your changes to your forked repository (usually called `origin`) and then initiating a pull request.

```
git push origin master
git push origin dev
```

Now follow the steps below to initiate a Pull request on GitHub.

1.  Navigate your browser to your forked repository: <https://github.com/YOUR_ACCOUNT/rancher.github.io.git>.
1.  Click the *Compare & pull request* button on the upper side of the forked repository.
1.  Enter a clear description for the changes you have made.
1.  Click *Send Pull Request*.

If you are asked to make modifications to your proposed changes, make the changes locally on your `dev` branch and push the changes. The Pull Request will be automatically updated.

### Cleaning up the Local Repository

You no longer need the `dev` branch after the changes have been committed into `upstream/master`. If you want to make additional documentation changes, restart the process with a new branch.

```
git checkout master
git branch -D dev
git push origin :dev
```
