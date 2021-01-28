#!/bin/bash

RUNTIME_ENV_VARIABLES=(
    SLACK_TOKEN="xoxb-274669747541-1692375673748-QVaNuoegFyTs61nupMnd1JrC"
    SLACK_CHANNEL="C01KRCWF8B1"
)

SET_ENV_VARS_TEXT=$(IFS=,; echo "${RUNTIME_ENV_VARIABLES[*]}")

gcloud functions deploy entry \
    --entry-point RelayToHTTP \
    --runtime go113 \
    --trigger-http \
    --allow-unauthenticated \
    --set-env-vars "${SET_ENV_VARS_TEXT}"
