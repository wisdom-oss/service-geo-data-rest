FROM python:3.10-slim
LABEL vendor="WISdoM 2.0 Project Group"
LABEL maintainer="wisdom@uol.de"
# Do not change this variable. Use the environment variables in docker compose or while starting to modify this value
ENV CONFIG_HTTP_PORT=5000
# TODO: Change this variable to the name of your service or use docker compose to set it
ENV CONFIG_SERVICE_NAME="geodata-rest-test"

WORKDIR /service
COPY . /service
RUN python -m pip install -r /service/requirements.txt
RUN python -m pip install gunicorn
RUN python -m pip install uvicorn[standard]
RUN ln ./configuration/gunicorn.py gunicorn.config.py
EXPOSE $CONFIG_HTTP_PORT
ENTRYPOINT ["gunicorn", "-cgunicorn.config.py", "api:service"]
