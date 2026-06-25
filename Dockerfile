FROM --platform=$BUILDPLATFORM docker.io/hrfee/jfa-go-build-docker:latest AS support
ARG BUILT_BY
ARG JFA_GO_VERSION=git
ARG JFA_GO_CSS_VERSION=v0.6.0
ARG JFA_GO_NFPM_EPOCH=0
ARG JFA_GO_BUILD_TIME
ARG NPM_CONFIG_REGISTRY=https://registry.npmjs.org/
ENV JFA_GO_BUILT_BY=$BUILT_BY
ENV JFA_GO_VERSION=$JFA_GO_VERSION
ENV JFA_GO_CSS_VERSION=$JFA_GO_CSS_VERSION
ENV JFA_GO_NFPM_EPOCH=$JFA_GO_NFPM_EPOCH
ENV JFA_GO_BUILD_TIME=$JFA_GO_BUILD_TIME
ENV CSSVERSION=$JFA_GO_CSS_VERSION
ENV BUILTBY=$BUILT_BY
ENV NPM_CONFIG_REGISTRY=$NPM_CONFIG_REGISTRY
ENV NPM_CONFIG_AUDIT=false
ENV NPM_CONFIG_FUND=false
ENV NPM_CONFIG_FETCH_RETRIES=5
ENV NPM_CONFIG_FETCH_RETRY_MINTIMEOUT=20000
ENV NPM_CONFIG_FETCH_RETRY_MAXTIMEOUT=120000
ENV NPM_CONFIG_FETCH_TIMEOUT=120000

COPY . /opt/build

RUN cd /opt/build; INTERNAL=off UPDATER=docker ./scripts/version.sh goreleaser build --snapshot --skip=validate --clean --id notray-e2ee
RUN mv /opt/build/dist/*_linux_arm_6 /opt/build/dist/placeholder_linux_arm
RUN sed -i 's#id="password_resets-watch_directory" placeholder="/config/jellyfin"#id="password_resets-watch_directory" value="/jf" disabled#g' /opt/build/build/data/html/setup.html

FROM gcr.io/distroless/base:latest AS final
ARG TARGETARCH

COPY --from=support /opt/build/dist/*_linux_${TARGETARCH}* /jfa-go
COPY --from=support /opt/build/build/data /jfa-go/data

EXPOSE 8056
EXPOSE 8057

CMD [ "/jfa-go/jfa-go", "-data", "/data" ]
