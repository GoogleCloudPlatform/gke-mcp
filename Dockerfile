# checkov:skip=CKV_DOCKER_2:Existing issue, suppressing to unblock presubmit
# checkov:skip=CKV_DOCKER_3:Existing issue, suppressing to unblock presubmit
FROM node:22-slim AS ui-build

WORKDIR /ui
COPY ui/package*.json ./
RUN npm ci
COPY ui/ ./
RUN npm run build

FROM golang:1.26.2 AS build

WORKDIR /go/src/gke-mcp
COPY go.mod go.sum ./
RUN go mod download

# Copy .git directory to preserve git information for versioning.
COPY .git ./.git
COPY *.go .
COPY cmd/ ./cmd/
COPY pkg/ ./pkg/
COPY ui/ui.go ./ui/ui.go
COPY --from=ui-build /ui/dist ./ui/dist/
RUN CGO_ENABLED=0 go build -o /gke-mcp

# Use the google-cloud-cli image as the base image because it contains the
# gcloud and kubectl binaries.
FROM gcr.io/google.com/cloudsdktool/google-cloud-cli:558.0.0-debian_component_based-20260224

COPY --from=build /gke-mcp /usr/local/bin/gke-mcp

EXPOSE 8080
ENTRYPOINT [ "gke-mcp" ]
