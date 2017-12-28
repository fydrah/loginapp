#!/bin/bash

./login-app --listen "http://0.0.0.0:5555" \
             --client-id minikube \
             --client-secret ZXhhbXBsZS1hcHAtc2VjcmV0 \
             --issuer https://dex.dex.local:5554/dex \
             --issuer-root-ca ca.crt \
             --redirect-uri http://127.0.0.1:5555/callback \
             --disable-choices \
             --extra-scopes groups \
             --app-name "Kubernetes login"
