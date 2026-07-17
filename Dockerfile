FROM public.ecr.aws/d3j8x8q7/olympus-base-go:latest

WORKDIR /app

# Copy dependency manifests first to cache the download layer separately from source.
COPY go.mod go.sum ./
COPY modules/postgres/go.mod modules/postgres/go.sum modules/postgres/

# Download all dependencies while network is available.
# modules/postgres/go.mod has a replace directive pointing to the root (../../..),
# so root dependencies must be downloaded first.
RUN go mod download && \
    cd modules/postgres && go mod download

# Copy full source and pre-compile to populate the build cache.
COPY . .

RUN go build ./... && \
    cd modules/postgres && go build ./...

CMD ["/bin/bash"]
