# iconikclient2
A golang client for Iconik's REST API. This is used for uploading, searching, and generating embeddable URLs of video assets.

# Tech stack

The client is written in Go. JSON objects for each of the relevant Iconik Models are written in `apitypes.go`. The Iconik Client and its methods are defined in `client.go`.

We expect new code to:

- Be reviewed by another team member on a pull request
- Have unit tests
- Be well-documented, clear, and modularized
- Be properly formatted with an auto-formatter (prettier, go fmt, etc.), preferably done on-save by your editor

# Getting started

## TLDR

Mac:

### Install Go Version Manager and Go
curl -sL -o /usr/local/bin/gvm https://github.com/andrewkroh/gvm/releases/download/v0.2.0/gvm-darwin-amd64
chmod +x /usr/local/bin/gvm
gvm 1.17.8 >> ~/.bashrc
eval "$(gvm 1.17.8)"
go version

Then clone the repository.

## Native Local Development Setup

### Setting up
You will need a token and an App-ID key that you can use. The instructions are found [here](https://app.iconik.io/docs/api.html#gettingstarted). You will either need admin access to an Iconik system, or ask an admin to generate the keys for you.

You can then create a main.go that constructs an Iconik Client struct and call relevant methods that you desire to test.

**DO NOT COMMIT YOUR MAIN.GO CONTAINING THE TOKEN/APP-ID key TO THE REPO**.

### Aim for simplicity

One goal with this stack is simplicity. Generally we avoid introducing new dependencies that will increase the learning curve for development.


### Deploying to Production

**DO NOT PUSH DIRECTLY TO MASTER**

Rather, create a seperate branch, do your commits on that branch, and then submit a [pull request](https://github.com/jzhang919/iconikclient2/pulls) to master once your feature is complete.
