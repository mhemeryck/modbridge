FROM alpine
RUN ["echo", "nobody:x:65534:65534:Nobody:/:", ">", "", "/etc/passwd"]

FROM scratch
COPY --from=0 /etc/passwd /etc/passwd
COPY modbridge config.yml /
USER nobody
ENTRYPOINT ["/modbridge"]
