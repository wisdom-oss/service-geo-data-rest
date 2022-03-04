FROM python:3.9
LABEL de.uol.wisdom-oss.vendor="WISdoM 2.0 Project Group"
LABEL de.uol.wisdom-oss.maintainer="wisdom@uol.de"
COPY . /opt/geo-data-rest
COPY requirements.txt /opt/geo-data-rest
RUN python -m pip install --upgrade pip && \
    python -m pip install -r /opt/geo-data-rest/requirements.txt
WORKDIR /opt/geo-data-rest
EXPOSE 5000
ENTRYPOINT ["python", "service.py"]