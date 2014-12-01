# TODO Tracks

TODO Tracks is a tool to let users get a handle on the various TODOs they or
their teammate have added over time. This allows people to track progress by
examining the TODOs remaining.

The tool will examine all the branches, local and remote and process the TODOs
in the different revisions and present them to the user. 

Use cases:

* List the TODOs in a branch.
* Examine when a TODO was added, removed, and who added it.
* (coming soon) Show which branches a TODO is in.
* (coming soon) Show a diff of the TODOs between two branches.

## Disclaimer

This is not an official Google product.

<!--
TODO: Add a getting started section.
-->
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

