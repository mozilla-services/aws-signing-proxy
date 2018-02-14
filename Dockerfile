FROM scratch
ADD aws-signing-proxy /aws-signing-proxy
ADD cacert.pem /
CMD ["/aws-signing-proxy"]
