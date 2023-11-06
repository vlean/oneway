FROM ubuntu:22.04



RUN set -x && apt-get update \
    && apt-get install --no-install-recommends ca-certificates -y \
    && update-ca-certificates \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

ENV PATH="/app:$PATH"

COPY bin/app /app

EXPOSE 80 443

WORKDIR /app

CMD ["oneway", "client"]