#!/usr/bin/env bash

docker run \
    -v "./adc.secret.json:/secrets/adc.json" \
    -p "9000:9000" \
    --name "marcy-webhooks" \
    -it "gcr.io/marcydotcloud/marcy-webhooks:nojson"