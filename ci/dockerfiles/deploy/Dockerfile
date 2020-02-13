FROM relintdockerhubpushbot/cf-deployment-concourse-tasks
MAINTAINER https://github.com/cloudfoundry/cf-k8s-networking

ENV HELM_VERSION 2.12.3
ENV KAPP_VERSION 0.19.0
ENV KAPP_CHECKSUM "99a2597d29ab9cf75d636a8220cb7e5ee315ac85b7adeb48b6c1ccb56a5cf477"
ENV YTT_VERSION 0.25.0
ENV YTT_CHECKSUM "2ccd2c74fea87748515ed3fd74cd45c31f91472fabadf42e6a958f52312e3b47"


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

# install kubectl
RUN apt-get install kubectl

# Get Helm
RUN wget https://storage.googleapis.com/kubernetes-helm/helm-v${HELM_VERSION}-linux-amd64.tar.gz -P /tmp && \
  tar -xvf /tmp/helm-v${HELM_VERSION}-linux-amd64.tar.gz -C /tmp && \
  mv /tmp/linux-amd64/helm /usr/local/bin/helm

# install k14s utils (ytt, kapp)
RUN  wget -O- https://github.com/k14s/ytt/releases/download/${YTT_VERSION}/ytt-linux-amd64 > /tmp/ytt && \
    echo "${YTT_CHECKSUM}  /tmp/ytt" | shasum -c - && \
    mv /tmp/ytt /usr/local/bin/ytt && \
    chmod +x /usr/local/bin/ytt && \

    wget -O- https://github.com/k14s/kapp/releases/download/${KAPP_VERSION}/kapp-linux-amd64 > /tmp/kapp && \
    echo "${KAPP_CHECKSUM}  /tmp/kapp" | shasum -c - && \
    mv /tmp/kapp /usr/local/bin/kapp && \
    chmod +x /usr/local/bin/kapp

# Clean up
RUN apt-get remove -y python-dev apt-transport-https && \
apt-get -y clean && apt-get -y autoremove --purge && rm -rf /etc/apt/ && \
rm -rf /tmp/* && \
find /var/lib/apt/lists -type f | xargs rm -f && \
find /var/cache/debconf -type f -name '*-old' | xargs rm -f && \
find /var/log -type f -user root | xargs rm -rf && \
for file in $(find /var/log -type f -user syslog); do echo > $file; done
