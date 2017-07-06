#!/bin/bash

if [ "$TRAVIS_PULL_REQUEST" == "false" ] && [ "$TRAVIS_BRANCH" == "master" ]; then
   docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD";
   docker build --no-cache -t $DOCKER_IMAGE_NAME .;
   docker tag $DOCKER_IMAGE_NAME $DOCKER_IMAGE_NAME:build-$TRAVIS_BUILD_NUMBER;
   docker tag $DOCKER_IMAGE_NAME $DOCKER_IMAGE_NAME:latest;
   docker push $DOCKER_IMAGE_NAME:build-$TRAVIS_BUILD_NUMBER;
   docker push $DOCKER_IMAGE_NAME:latest;
fi
