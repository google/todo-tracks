#!/bin/bash

# Copyright 2014 Google Inc. All rights reserved.
# 
# 	Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
# http://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Helper script for launching the TODO tracker inside of a GCE VM.
# This needs to be run using root privileges

# First set up the VM to have the appropriate pre-requisites
#   Install git from backport, base wheezy git is 1.7.x which isn't
apt-get -qqy update
apt-get -qqy -t wheezy-backports install git
#   Update gcloud to have the preview components (which includes deployment-manager)
gcloud --quiet components update
gcloud --quiet components update preview

# Next, initialize/clone git repo.
mkdir -p repo && cd repo
export PROJECT=$(curl http://metadata.google.internal/computeMetadata/v1/project/project-id -H "Metadata-Flavor: Google")
gcloud init ${PROJECT}
#   Configure cronjob to pull git repo every minute.
echo "* * * * * su -s /bin/sh root -c 'cd $(pwd)/${PROJECT}/default && /usr/bin/git pull'" >> /tmp/crontab.txt
crontab /tmp/crontab.txt

# Finally, copy over todo binary from Google Cloud Storage bucket, and set it up.
mkdir -p bin && wget http://storage.googleapis.com/todo-track-bin/todos -O bin/todos
#   Set executable bit on todo binary.
chmod +x bin/todos
#   Start running todo server.
cd ${PROJECT}/default; ../../bin/todos --todo_regex="TODO([(]|:)" &

