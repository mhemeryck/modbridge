FROM scratch
COPY --from=modbridge:build /app ./
ENTRYPOINT ["./app"]
