FROM alpine:3.7
COPY ./istiofilter /bin/istiofilter
ENTRYPOINT ["/bin/istiofilter"]