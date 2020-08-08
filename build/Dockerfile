FROM registry.access.redhat.com/ubi8-minimal
MAINTAINER Enlin Xu <enlin.xu@turbonomic.com>
ARG GIT_COMMIT
ENV GIT_COMMIT ${GIT_COMMIT}

# Required OpenShift Labels
LABEL name="Turbodif Container" \
      vendor="Turbonomic" \
      version="v7.22.0" \
      release="1" \
      summary="Performance assurance for the applications in Openshift" \
      description="Hybrid Cloud Container leverages Turbonomic control platform, to assure the performance of micro-services running in OpenShift, as well as the efficiency of underlying infrastructure." \
### Required labels above - recommended below
      url="https://www.turbonomic.com" \
      run='docker run -tdi --name ${NAME} turbonomic/turbodif:latest' \
      io.k8s.description="Turbonomic Workload Automation Platform simultaneously optimizes performance, compliance, and cost in real-time. Workloads are precisely resourced, automatically, to perform while satisfying business constraints.  " \
      io.k8s.display-name="Turbodif Operator" \
      io.openshift.expose-services="" \
      io.openshift.tags="turbonomic, Multicloud Container"

### add licenses to this directory
COPY licenses /licenses

COPY Dockerfile /Dockerfile

### Setup user for build execution and application runtime
ENV APP_ROOT=/opt/turbonomic
ENV PATH=$PATH:${APP_ROOT}/bin

RUN mkdir -p ${APP_ROOT}/bin
RUN mkdir -p ${APP_ROOT}/conf
COPY build/turbodif.linux ${APP_ROOT}/bin/turbodif
COPY configs/app-supply-chain-config.yaml ${APP_ROOT}/conf/app-supply-chain-config.yaml
RUN chmod -R ug+x ${APP_ROOT}/bin && sync && \
    chmod -R g=u ${APP_ROOT}

####### Add app-specific needs below. #######
USER 10001
WORKDIR ${APP_ROOT}
ENTRYPOINT ["/opt/turbonomic/bin/turbodif"]
