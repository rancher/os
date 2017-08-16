.PHONY: default build build-nginx run live

default: build build-nginx run

# this target sets up Jekyll, and uses it to build html
build:
	docker build -t rancher/rancher.github.io:build -f Dockerfile.build .

# this target uses the Jekyll html image in a multistage Dockerfile to build a small nginx image
build-nginx:
	docker build -t rancher/rancher.github.io .

run: build-nginx
	docker run --rm -p 80:80 rancher/rancher.github.io

# this target will use the jekyll image and bind mount your local repo, when you modify a file, the html will be automatically rebuilt. (the redirects from latest won't work)
# You can also examine the output html in the _sites dir.
live:
	docker run --rm -it -p 80:4000 -v $(PWD):/build rancher/rancher.github.io:build jekyll serve -w -P 4000 --incremental

clean:
	rm -rf _sites
	docker rmi rancher/rancher.github.io
	docker rmi rancher/rancher.github.io:build
