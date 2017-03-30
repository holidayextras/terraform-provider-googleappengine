# terraform-provider-googleappengine
This provider exposes resources that we don't want to build into
terraform mainline.  There are two reasons for this:
1. we are moving very quickly right now
2. the scope of adding a real appengine provider is massive and
   see number 1

This provider only supports java applications right now.  Any others
will fails.

## Current Version
Terraform are not easy to work with in terms of plugins. Currently their
plugin API is version 4. This plugin has jumped from 2 to 4. There are
breaking changes in between terraform version, up and down. This plugin can
only be run with these specifics in mind:
* terraform version `0.8.8`
* gcloud version `146.0.0`
* go version `1.8`

## CI Deployment
With [dep](https://github.com/golang/dep) as the package manager we can
create vendor on deploy instead of carrying it around with us in the repo.
CircleCI caching will alleviate some of the download time. Dep allows us to
slim the repo considerably by removing the vendor and creating it on deploy.

Dep is in alpha but works currently, it will be the official dependency
manager for go, so integrating with that early.

To use:
- check out
- run tests 
  - add several variables to your environment:
    - TF_ACC (set to anything)
    - GOOGLE_CREDENTIALS, set to the contents of a secrets file downloaded from google
    - GOOGLE_PROJECT, set to the project for the above credentials
    - GOOGLE_REGION, set to us-central1
  - execute tests using makefile
    - make tests TESTARGS=<args to pass to 'go test'>
    - ex to only run dataflow tests "make test TESTARGS='--run=Dataflow'"
- install binary to $GOBIN to make it usable system wide (assumes GOBIN is in your PATH)
  - make install
- edit terraform.rc (see terraform docs here: https://terraform.io/docs/plugins/basics.html) to have the
  following block:
  providers {
    googleappengine = "terraform-provider-googleappengine"
  }
- build and copy file to terraform install
  - locate terraform install
  - go build -o TERRAFORM_INSTALL_LOCATION

