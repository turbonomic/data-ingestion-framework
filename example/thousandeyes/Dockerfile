FROM registry.access.redhat.com/ubi8-minimal
MAINTAINER Meng Ding <meng.ding@turbonomic.com>
ARG GIT_COMMIT
ENV GIT_COMMIT ${GIT_COMMIT}

### Setup user for build execution and application runtime
ENV APP_ROOT=/opt/turbonomic
ENV PATH=$PATH:${APP_ROOT}/bin

RUN mkdir -p ${APP_ROOT}/bin
COPY thousandeyes.linux ${APP_ROOT}/bin/thousandeyes
RUN chmod -R ug+x ${APP_ROOT}/bin && sync && \
    chmod -R g=u ${APP_ROOT}

####### Add app-specific needs below. #######
USER 10001
WORKDIR ${APP_ROOT}
ENTRYPOINT ["/opt/turbonomic/bin/thousandeyes"]
