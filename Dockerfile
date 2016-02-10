FROM node

# kocho
COPY kocho /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/kocho"]
