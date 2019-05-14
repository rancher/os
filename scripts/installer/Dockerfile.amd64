FROM rancher/os-installer

# TODO: separate out the elements below - so we can mix and match updates
RUN rm /dist/ \
    && mkdir -p /dist/

COPY ./boot/ /dist/
