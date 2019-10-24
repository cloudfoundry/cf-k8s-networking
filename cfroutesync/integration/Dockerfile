# gcr.io/cf-routing/cf-k8s-networking/cfroutesync-integration-test-env
FROM cloudfoundry/cflinuxfs3

ENV GO_VERSION 1.13.1

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
            openssl \
            vim \
            zip \
            python-pip \
            zlib1g-dev \
      && \
      apt clean

# Install go
RUN curl "https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz" | tar xvz -C /usr/local/

# Add golang environment variables
ENV HOME /root
ENV GOPATH /root/go
ENV PATH "${PATH}:/usr/local/go/bin:/root/bin:${GOPATH}/bin"

# Install ginkgo
RUN go get github.com/onsi/ginkgo/ginkgo


# Install kubebuilder
RUN \
  curl -L -O "https://github.com/kubernetes-sigs/kubebuilder/releases/download/v1.0.8/kubebuilder_1.0.8_linux_amd64.tar.gz" && \
  tar -zxvf kubebuilder_1.0.8_linux_amd64.tar.gz > /dev/null && \
  mv kubebuilder_1.0.8_linux_amd64 /usr/local/kubebuilder

ENV PATH="${PATH}:/usr/local/kubebuilder/bin"

# Run metacontroller in this docker image
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

RUN curl -L -O  "https://github.com/GoogleCloudPlatform/metacontroller/archive/v0.4.0.tar.gz" && \
    tar -zxvf v0.4.0.tar.gz > /dev/null && \
    mv metacontroller-0.4.0 ${GOPATH}/src/metacontroller.app/ && \
    cd ${GOPATH}/src/metacontroller.app/ && \
    dep ensure && go install && \
    mv ${GOPATH}/bin/metacontroller.app /usr/local/bin/metacontroller

# Install galley for the validating admission controller
COPY --from=docker.io/istio/galley:1.3.3 /usr/local/bin/galley /usr/local/bin/galley


