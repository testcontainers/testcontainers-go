FROM public.ecr.aws/d3j8x8q7/olympus-base-go:latest

WORKDIR /app

COPY . .

# Download dependencies for the root module and the postgres module.
# modules/postgres/go.mod has a replace directive pointing to the root (../../..),
# so the root module must be present and downloaded first for offline compilation.
RUN go mod download && \
    cd modules/postgres && go mod download

CMD ["/bin/bash"]
