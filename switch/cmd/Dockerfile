FROM alpine:3.4
WORKDIR /srv
ADD third_party third_party
ARG DEPLOY_ENV=local
ENV DEPLOY_ENV=${DEPLOY_ENV}
COPY cmd/device/configs/ configs/device/ 
COPY cmd/master/configs/ configs/master/
COPY cmd/slave/configs/ configs/slave/
ADD ui ui
ADD build/linux_amd64/switch .
