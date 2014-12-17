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

FROM ubuntu

ENV DEBIAN_FRONTEND noninteractive
RUN apt-get update && apt-get upgrade -y && \
  apt-get install -y -qq --no-install-recommends \
  ca-certificates \
  git \
  curl \
  python \
  unzip

ADD http://storage.googleapis.com/todo-track-bin/todos /bin/todos
RUN chmod +x /bin/todos

ADD https://dl.google.com/dl/cloudsdk/release/google-cloud-sdk.zip /google-cloud-sdk.zip
RUN unzip /google-cloud-sdk.zip -d /
RUN rm /google-cloud-sdk.zip
RUN /google-cloud-sdk/install.sh --usage-reporting=true --path-update=true --bash-completion=true --rc-path=/.bashrc --disable-installation-options
ENV PATH /google-cloud-sdk/bin:$PATH

ADD gce.sh /bin/gce.sh
RUN chmod +x /bin/gce.sh
RUN touch /var/log/cron.log

EXPOSE 8080
CMD ["/bin/sh", "-c", "/bin/gce.sh & cron; tail -f /var/log/cron.log"]
