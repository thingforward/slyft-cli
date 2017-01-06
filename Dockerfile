FROM debian:jessie

ENV DEBIAN_FRONTEND noninteractive
RUN mkdir /slyft-cli /slyft-cli-build
WORKDIR /slyft-cli
ENV GOPATH=/slyft-cli-build/.golang
ENV GOBIN=$GOPATH/bin
RUN mkdir $GOPATH $GOBIN
RUN apt-get update -y -q \
	&& apt-get upgrade -y -q \
	&& apt-get install -y -q golang nodejs npm git zip \
	&& cd /slyft-cli-build && git clone https://github.com/thingforward/slyft-cli.git \
	&& cd slyft-cli \
	&& npm install \
	&& nodejs node_modules/gulp/bin/gulp.js \
	&& mv dist/* /slyft-cli/ \
	&& mv bin/* /slyft-cli/ \
	&& rm -fr /slyft-cli-build \
    	&& apt-get clean -y -q \
    	&& apt-get autoclean -y -q \
    	&& apt-get autoremove -y -q \
    	&& rm -rf /usr/share/locale/* \
    	&& rm -rf /var/cache/debconf/*-old \
    	&& rm -rf /var/lib/apt/lists/* \
    	&& rm -rf /usr/share/doc/*

CMD [ "/slyft-cli/slyft-cli" ]
