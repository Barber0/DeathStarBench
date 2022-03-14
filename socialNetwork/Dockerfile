FROM simonalphafang/socialnetwork-env:0.0.2

COPY ./ /social-network-microservices
RUN cd /social-network-microservices \
    && mkdir -p build \
    && cd build \
    && cmake -DCMAKE_BUILD_TYPE=Debug .. \
    && make -j2 \
    && make install

WORKDIR /social-network-microservices