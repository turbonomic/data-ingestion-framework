FROM python:3
MAINTAINER Meng Ding <meng.ding@turbonomic.com>
ARG GIT_COMMIT
ENV GIT_COMMIT ${GIT_COMMIT}

### Setup dependencies
COPY requirements.txt /tmp/
RUN pip install --requirement /tmp/requirements.txt

### Setup user for build execution and application runtime
ENV APP_ROOT=/opt/turbonomic
ENV PATH=$PATH:${APP_ROOT}/bin

### Copy executables
RUN mkdir -p ${APP_ROOT}/bin
COPY azureloganalytics.py ${APP_ROOT}/bin/
COPY dif.py ${APP_ROOT}/bin/
RUN chmod -R ug+x ${APP_ROOT}/bin && sync && \
    chmod -R g=u ${APP_ROOT}

####### Add app-specific needs below. #######
USER 10001
WORKDIR ${APP_ROOT}
CMD ["/opt/turbonomic/bin/azureloganalytics.py"]
ENTRYPOINT ["python"]
