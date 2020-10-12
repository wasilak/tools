FROM --platform=$TARGETPLATFORM python:3-slim

ARG TARGETPLATFORM
ARG BUILDPLATFORM

RUN echo "I am running on $BUILDPLATFORM, building for $TARGETPLATFORM" > /log

# RUN pip install -U wheel

RUN apt-get update && apt install build-essential -y

COPY ./requirements.txt /requirements.txt

RUN pip install -r requirements.txt

COPY ./app /app/

WORKDIR /app

EXPOSE 5000

ENV SECRET_KEY=cnfg8237hc38aadsadaigochh98cy^TR^&%R&T*&G

ENV SESSION_COOKIE_NAME=session-fastapi-tools

ENTRYPOINT ["uvicorn", "main:app", "--host=0.0.0.0", "--port=5000"]

CMD ["--reload", "--log-level=info"]
