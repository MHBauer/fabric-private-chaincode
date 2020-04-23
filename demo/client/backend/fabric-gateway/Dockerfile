#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#

FROM golang:1.14-buster



COPY \
    config.json\
    auction.json\
  /usr/src/app/
COPY gateway /usr/src/app/gateway

WORKDIR /usr/src/app
COPY . .

RUN go build -v -o server ./src/

CMD ["./server"]