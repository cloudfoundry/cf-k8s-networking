FROM cloudfoundry/cflinuxfs3
MAINTAINER https://github.com/cloudfoundry/cf-k8s-networking

ENV BOSH_CLI_VERSION 5.4.0
ENV TERRAFORM_VERSION 0.11.4
ENV BBL_VERSION 6.9.16
ENV CF_CLI_VERSION 6.38.0
ENV CREDHUB_CLI_VERSION 2.0.0
ENV HELM_VERSION 2.12.3
ENV GO_VERSION 1.13.2

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

# Get cf
RUN \
  curl -sL "https://packages.cloudfoundry.org/stable?release=linux64-binary&version=${CF_CLI_VERSION}&source=github-rel" | tar zxf - -C /usr/bin

# Get credhub
RUN \
  wget https://github.com/cloudfoundry-incubator/credhub-cli/releases/download/${CREDHUB_CLI_VERSION}/credhub-linux-${CREDHUB_CLI_VERSION}.tgz -P /tmp && \
  tar -xvf /tmp/credhub-linux-${CREDHUB_CLI_VERSION}.tgz -C /tmp && \
  mv /tmp/credhub /usr/local/bin/credhub && \
  chmod +x /usr/local/bin/credhub

# Get terraform
RUN \
  wget https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip -P /tmp && \
  unzip /tmp/terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /usr/local/bin/

# Get BBL
RUN \
  wget https://github.com/cloudfoundry/bosh-bootloader/releases/download/v${BBL_VERSION}/bbl-v${BBL_VERSION}_linux_x86-64 -P /tmp && \
  mv /tmp/bbl-* /usr/local/bin/bbl && \
  chmod +x /usr/local/bin/bbl

# Get BOSH
RUN \
  wget https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-${BOSH_CLI_VERSION}-linux-amd64 && \
  chmod +x bosh-cli-${BOSH_CLI_VERSION}-linux-amd64 && \
  mv bosh-cli-${BOSH_CLI_VERSION}-linux-amd64 /usr/local/bin/bosh2 && \
  cp /usr/local/bin/bosh2 /usr/local/bin/bosh

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

# Install go
RUN curl "https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz" | tar xvz -C /usr/local/

# Add golang environment variables
ENV HOME /root
ENV GOPATH /root/go
ENV PATH "${PATH}:/usr/local/go/bin:/root/bin:${GOPATH}/bin"

# Clean up
RUN apt-get remove -y python-dev apt-transport-https && \
apt-get -y clean && apt-get -y autoremove --purge && rm -rf /etc/apt/ && \
rm -rf /tmp/* && \
find /var/lib/apt/lists -type f | xargs rm -f && \
find /var/cache/debconf -type f -name '*-old' | xargs rm -f && \
find /var/log -type f -user root | xargs rm -rf && \
for file in $(find /var/log -type f -user syslog); do echo > $file; done
