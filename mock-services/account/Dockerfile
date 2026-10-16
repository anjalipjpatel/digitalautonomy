FROM node:20-alpine

RUN npm install -g @mockoon/cli@9.9.0
COPY ./account-mockoon.json ./account-mockoon.json

# Do not run as root.
RUN adduser --shell /bin/sh --disabled-password --gecos "" mockoon
RUN chown -R mockoon ./account-mockoon.json
USER mockoon

EXPOSE 3000

ENTRYPOINT ["mockoon-cli","start","--disable-log-to-file","--data","./account-mockoon.json","--port","3000"]

# Usage: docker run -p <host_port>:3000 mockoon-test