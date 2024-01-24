# TODO Tracks

TODO Tracks is a tool to let users get a handle on the various TODOs they or
their teammate have added over time. This allows people to track progress by
examining the TODOs remaining.

The tool examines all the branches in a git repo (local and remote), finds the TODOs
in the different revisions, and presents them to the user. 

Use cases:

* List the TODOs in a branch.
* Examine when a TODO was added, removed, and who added it.
* Show which branches a TODO is in.

## Disclaimer

This is not an official Google product.

## Prerequisites

Building requires the Go tools and GNU Make. Running the built binary requires the git command line tool.

## Building the source code

First checkout the code from the git repo:

    git clone git@github.com:google/todo-tracks.git

Build the binary:

    make

And then launch it:

    bin/todos

The tracker requires that it be started in a directory that contains at least one git repo, and it shows the TODOs from every git repo under that directory.

The UI for the tracker is a webserver which defaults to listening on port 8080. To use a different port, pass it as an argument to the "--port" flag:

    bin/todos --port=12345

For more details about the supported command line flags, pass in the "--help" flag.

    bin/todos --help
