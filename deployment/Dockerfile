FROM golang:1.16.5 as builder
ADD . /go/xenotification
WORKDIR  /go/xenotification

COPY storage/gitlab_private_ssh /root/.ssh/id_rsa
RUN chmod 400 /root/.ssh/id_rsa
RUN eval "$(ssh-agent)" && ssh-add /root/.ssh/id_rsa
RUN ssh-keyscan -H gitlab.revenuemonster.my >> ~/.ssh/known_hosts
RUN git config --global url."git@gitlab.revenuemonster.my:".insteadOf "https://gitlab.revenuemonster.my/"

RUN make

FROM golang:1.16.5-alpine3.13
COPY --from=builder /go/xenotification /go/app/xenotification

WORKDIR  /go/app/xenotification
ENTRYPOINT ./xenotification