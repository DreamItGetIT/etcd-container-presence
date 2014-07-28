FROM ubuntu:12.04

RUN apt-get update --fix-missing
RUN apt-get upgrade -y
ADD register /bin/register

ENTRYPOINT ["/bin/register"]
