FROM relintdockerhubpushbot/cf-deployment-concourse-tasks
MAINTAINER https://github.com/cloudfoundry/cf-k8s-networking

ENV HELM_VERSION 2.12.3

RUN \
      apt update && \
      apt -y install --fix-missing \
            build-essential \
            curl \
            docker \
            git \
            netcat-openbsd \
            htop \
            libpython-dev \
            lsof \
            psmisc \
            python \
            strace \
            libreadline6-dev\
            lsb-core \
            wget \
            unzip \
            libfontconfig1-dev libfreetype6 libssl-dev libpng-dev libjpeg-dev \
            jq \
            libssl-dev \
            libssl1.0.0 \
            libxml2-dev \
            libxslt-dev \
            libyaml-dev \
            ruby-all-dev \
            openssl \
            vim \
            zip \
            python-pip \
            zlib1g-dev \
      && \
      apt clean


# Temp fix to get around apt-key issues with canonical
RUN chmod 1777 /tmp

# gcloud SDK
RUN export CLOUD_SDK_REPO="cloud-sdk-$(lsb_release -c -s)" && \
    echo "deb http://packages.cloud.google.com/apt $CLOUD_SDK_REPO main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list && \
    curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add - && \
    apt-get update -y && apt-get install google-cloud-sdk -y

# install kubectl using gcloud
RUN apt-get install kubectl

# Get Helm
RUN wget https://storage.googleapis.com/kubernetes-helm/helm-v${HELM_VERSION}-linux-amd64.tar.gz -P /tmp && \
  tar -xvf /tmp/helm-v${HELM_VERSION}-linux-amd64.tar.gz -C /tmp && \
  mv /tmp/linux-amd64/helm /usr/local/bin/helm

# Clean up
RUN apt-get remove -y python-dev apt-transport-https && \
apt-get -y clean && apt-get -y autoremove --purge && rm -rf /etc/apt/ && \
rm -rf /tmp/* && \
find /var/lib/apt/lists -type f | xargs rm -f && \
find /var/cache/debconf -type f -name '*-old' | xargs rm -f && \
find /var/log -type f -user root | xargs rm -rf && \
for file in $(find /var/log -type f -user syslog); do echo > $file; done
