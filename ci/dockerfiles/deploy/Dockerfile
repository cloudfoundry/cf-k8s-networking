FROM relintdockerhubpushbot/cf-deployment-concourse-tasks
MAINTAINER https://github.com/cloudfoundry/cf-k8s-networking

ENV HELM_VERSION 2.12.3
ENV KAPP_VERSION "v0.24.0"
ENV KAPP_CHECKSUM "044a8355c1a3aa4c9e427fc64f7074b80cb759e539771d70d38933886dbd2df4"
ENV YTT_VERSION "v0.26.0"
ENV YTT_CHECKSUM "ea740f72a1825c00eabf83bfd6cb366adf4b2486992603d6850fac8487e92d8f"
ENV ISTIOCTL_VERSION 1.4.3

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

# cf cli
RUN wget -q -O - https://packages.cloudfoundry.org/debian/cli.cloudfoundry.org.key | apt-key add - && \
    echo "deb https://packages.cloudfoundry.org/debian stable main" | tee /etc/apt/sources.list.d/cloudfoundry-cli.list && \
    apt-get update -y && apt-get install cf-cli -y

# gcloud SDK
RUN export CLOUD_SDK_REPO="cloud-sdk-$(lsb_release -c -s)" && \
    echo "deb http://packages.cloud.google.com/apt $CLOUD_SDK_REPO main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list && \
    curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add - && \
    apt-get update -y && apt-get install google-cloud-sdk -y

# install kubectl
RUN apt-get install kubectl

# install istioctl
RUN tmp=$(mktemp -d /tmp/istioctl.XXXXXX) && \
  filename="istioctl-${ISTIOCTL_VERSION}-linux.tar.gz" && \
  cd "$tmp" && \
  URL="https://github.com/istio/istio/releases/download/${ISTIOCTL_VERSION}/istioctl-${ISTIOCTL_VERSION}-linux.tar.gz" && \
  curl -sLO "${URL}" &&\
  tar -xzf "${filename}" && \
  mv "${tmp}/istioctl" /usr/local/bin/istioctl && \
  chmod +x /usr/local/bin/istioctl && \
  rm -r "${tmp}"

# Get Helm
RUN wget https://storage.googleapis.com/kubernetes-helm/helm-v${HELM_VERSION}-linux-amd64.tar.gz -P /tmp && \
  tar -xvf /tmp/helm-v${HELM_VERSION}-linux-amd64.tar.gz -C /tmp && \
  mv /tmp/linux-amd64/helm /usr/local/bin/helm

# install k14s utils (ytt, kapp)
RUN  wget -O- https://github.com/k14s/ytt/releases/download/${YTT_VERSION}/ytt-linux-amd64 > /tmp/ytt && \
    echo "${YTT_CHECKSUM}  /tmp/ytt" | shasum -c - && \
    mv /tmp/ytt /usr/local/bin/ytt && \
    chmod +x /usr/local/bin/ytt

RUN wget -O- https://github.com/k14s/kapp/releases/download/${KAPP_VERSION}/kapp-linux-amd64 > /tmp/kapp && \
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
