# TODO Tracks

[![Build Status](https://travis-ci.org/google/todo-tracks.svg?branch=master)](https://travis-ci.org/google/todo-tracks)

TODO Tracks is a tool to let users get a handle on the various TODOs they or
their teammate have added over time. This allows people to track progress by
examining the TODOs remaining.

The tool examines all the branches in a git repo (local and remote), finds the TODOs
in the different revisions, and presents them to the user. 

Use cases:

* List the TODOs in a branch.
* Examine when a TODO was added, removed, and who added it.
* Show which branches a TODO is in.
* (coming soon) Show a diff of the TODOs between two branches.

## Disclaimer

This is not an official Google product.

<!--
TODO: Add a getting started section for running from a pre-built binary.
-->

## Prerequisites

Building requires the Go tools and GNU Make. Running the built binary requires the git command line tool.

## Building the source code

First checkout the code from the git repo:

    git clone git@github.com:google/todo-tracks.git

Build the binary:

    make

And then launch it:

    bin/todos

<!--
TODO(ojarjur): Add support for hg repos.
-->
The tracker requires that it be started in a directory that contains at least one git repo, and it shows the TODOs from every git repo under that directory.

The UI for the tracker is a webserver which defaults to listening on port 8080. To use a different port, pass it as an argument to the "--port" flag:

    bin/todos --port=12345

For more details about the supported command line flags, pass in the "--help" flag.

    bin/todos --help

<!--
TODO: Add a section detailing a sample workflow.
-->

## Running in Google Compute Engine

We provide a pre-built binary and config files for deploying the tool to a GCE VM
using Google Deployment Manager.

Assuming you already have the gcloud preview commands installed, run the following steps:

    mkdir todo-tracks
    cd todo-tracks
    wget http://storage.googleapis.com/todo-track-bin/config/gce.sh -O gce.sh
    wget http://storage.googleapis.com/todo-track-bin/config/gce.yaml -O gce.yaml
    gcloud preview deployment-manager templates create todo_tracks_template --template-file gce.yaml
    gcloud preview deployment-manager deployments --region=us-central1 create --template=todo_tracks_template todo_tracks_deployment

