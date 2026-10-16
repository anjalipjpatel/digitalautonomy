FROM node:20-alpine

RUN npm install -g @mockoon/cli@9.9.0
COPY ./payment-mockoon.json ./payment-mockoon.json

# Do not run as root.
RUN adduser --shell /bin/sh --disabled-password --gecos "" mockoon
RUN chown -R mockoon ./payment-mockoon.json
USER mockoon

EXPOSE 3000

ENTRYPOINT ["mockoon-cli","start","--disable-log-to-file","--data","./payment-mockoon.json","--port","3000"]

# Usage: docker run -p <host_port>:3000 mockoon-test