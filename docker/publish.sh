#!/bin/bash
set -e

tag=ngekaworu/user-center-go

docker build --file ./Dockerfile --tag ${tag} ..;
docker push ${tag};
