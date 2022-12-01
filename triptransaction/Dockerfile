FROM 950742359761.dkr.ecr.eu-west-1.amazonaws.com/base-docker-images:go-1.16


# TODO: remove this and test
ENV AWS_REGION=eu-west-1
ENV APP_ENVIRONMENT=staging

WORKDIR /go/src/app
COPY . $WORKDIR

RUN go mod download &&\
     bash build.sh


ENV PATH="/go/src/app/build/linux:$PATH"

CMD [ "trip-transaction", "--appName", "synadia", "--targetEnv" ,"staging" ]