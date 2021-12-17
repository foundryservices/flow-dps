FROM golang:1.17-buster AS build-setup

RUN apt-get update
RUN apt-get -y install cmake zip sudo git

#ENV FLOW_DPS_REPO="https://github.com/optakt/flow-dps"
#ENV FLOW_DPS_REPO="https://github.com/dapperlabs/flow-dps"
# v1.4.4 with some flag changes
#ENV FLOW_DPS_BRANCH=v0.23
#ENV FLOW_DPS_BRANCH=v1.4.4

#ENV DPS_ROSETTA_DOCKER_REPO="https://github.com/dapperlabs/dps-rosetta-docker"
#ENV DPS_ROSETTA_DOCKER_BRANCH=v0.23

#ENV FLOW_ROSETTA_REPO="https://github.com/dapperlabs/flow-rosetta"
#ENV FLOW_ROSETTA_BRANCH=wait

ENV FLOW_GO_REPO="https://github.com/onflow/flow-go"
ENV FLOW_GO_BRANCH=v0.23.3

#ENV ROSETTA_DISPATCHER_BRANCH=master

RUN mkdir /dps /docker /flow-go

WORKDIR /dps

# clone repos and create links
ADD . /dps
#RUN git clone --branch $DPS_ROSETTA_DOCKER_BRANCH $DPS_ROSETTA_DOCKER_REPO /docker
RUN git clone --branch $FLOW_GO_BRANCH $FLOW_GO_REPO /flow-go
#RUN git clone --branch $FLOW_ROSETTA_BRANCH $FLOW_ROSETTA_REPO /flow-rosetta

RUN ln -s /flow-go /dps/flow-go
#RUN ln -s /flow-go /flow-rosetta/flow-go

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build  \
    make -C /flow-go crypto/relic/build #prebuild crypto dependency

FROM build-setup AS build-binary

ARG BINARY

WORKDIR /dps
RUN  --mount=type=cache,target=/go/pkg/mod \
     --mount=type=cache,target=/root/.cache/go-build  \
     go build -o /app -tags relic -ldflags "-extldflags -static" ./cmd/$BINARY && \
     chmod a+x /app


## Add the statically linked binary to a distroless image
FROM gcr.io/distroless/base-debian11  AS production

ARG BINARY

COPY --from=build-binary /app /app

EXPOSE 5005

ENTRYPOINT ["/app"]