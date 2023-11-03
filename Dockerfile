FROM ubuntu:22.04

COPY bin/app /app

# RUN apt-get update \
#     && apt-get install ca-certificates -y \
#     && update-ca-certificates \
#     && apt-get clean \
#     && rm -rf /var/lib/apt/lists/*

ENV PATH="/app:$PATH"

EXPOSE 80 443

WORKDIR /app

CMD ["oneway", "client"]