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

# Initialize/clone git repo.
echo "Cloning the project repo"
mkdir -p repo && cd repo
export PROJECT=$(curl http://metadata.google.internal/computeMetadata/v1/project/project-id -H "Metadata-Flavor: Google")
gcloud init ${PROJECT}

echo "Configuring the cron job"
#   Configure cronjob to pull git repo every minute.
echo "* * * * * su -s /bin/sh root -c 'cd $(pwd)/${PROJECT}/default && /usr/bin/git pull'" >> /tmp/crontab.txt
crontab /tmp/crontab.txt

echo "Starting the TODO Tracks server"
#   Start running todo server.
cd ${PROJECT}/default; /bin/todos

