FROM haskell:9.6-slim-bullseye AS build

# Bullseye's own libpq-dev (13.x) is too old for postgresql-libpq-configure
# (requires libpq >= 14.12), so pull a current libpq from the PGDG apt repo.
RUN apt-get update && apt-get install -y --no-install-recommends wget gnupg ca-certificates && \
    install -d /usr/share/postgresql-common/pgdg && \
    wget -O /usr/share/postgresql-common/pgdg/apt.postgresql.org.asc --no-check-certificate https://www.postgresql.org/media/keys/ACCC4CF8.asc && \
    echo "deb [signed-by=/usr/share/postgresql-common/pgdg/apt.postgresql.org.asc] https://apt.postgresql.org/pub/repos/apt bullseye-pgdg main" > /etc/apt/sources.list.d/pgdg.list && \
    apt-get update && \
    apt-get install -y --no-install-recommends libpq-dev pkg-config gcc && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /build

COPY comments.cabal ./
RUN cabal update && cabal build --only-dependencies -j

COPY app ./app
COPY LICENSE CHANGELOG.md ./
RUN cabal build && \
    find dist-newstyle -type f -name comments -executable -exec cp {} /build/comments-bin \;

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends libpq5 ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=build /build/comments-bin /app/comments-bin

EXPOSE 8081
CMD ["/app/comments-bin"]
