ARG POSTGRES_VERSION
FROM postgres:${POSTGRES_VERSION}-alpine

RUN apk add --no-cache --virtual build-deps clang gcc git icu-dev libc-dev llvm make && \
    git clone "https://github.com/okbob/plpgsql_check" && cd plpgsql_check && make && make install && cd && \
    rm -rf plpgsql_check && apk del build-deps

COPY /init.sh /docker-entrypoint-initdb.d/init.sh
CMD [ "postgres", "-c", "shared_preload_libraries=plpgsql,pg_stat_statements,plpgsql_check", "-c", "plpgsql_check.profiler=on" ]