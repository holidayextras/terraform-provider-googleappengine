# terraform-provider-googleappengine
This provider exposes resources that we don't want to build into
terraform mainline.  There are two reasons for this:
1. we are moving very quickly right now
2. the scope of adding a real appengine provider is massive and
   see number 1

This provider only supports java applications right now.  Any others
will fails.

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
  - make build
  - cp terraform-provider-googleappengine TERRAFORM_INSTALL_LOCATION

And there is a Makefile that will do all of the above.  Makefile instructions:
To install:
  make install
  
To test:
  make test
  * this still requires you set the listed environment variables
