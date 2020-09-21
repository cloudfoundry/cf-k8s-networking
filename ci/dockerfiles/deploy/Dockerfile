FROM relintdockerhubpushbot/cf-for-k8s-ci
MAINTAINER https://github.com/cloudfoundry/cf-k8s-networking

RUN apk add --update-cache && \
  apk add --update alpine-sdk && \
  apk add libffi-dev openssl-dev && \
  apk --no-cache --update add build-base
