# slyft-cli
command line client to slyft-server

## Run slyft-cli
To run `slyft-cli`, fetch the appropriate zip file from the folder [dist](dist), extract and run:
```
$ ./slyft-cli
```

## Build slyft-cli
Before you begin, make sure you have Golang and Node.js installed. For the Go sources to build successfully, you also need $GOPATH and $GOBIN to be set (for this example, $GOPATH is set to ~/golang):
```
 $ cd
~$ mkdir -p golang/bin
~$ export GOPATH=~/golang
~$ export GOBIN=$GOPATH/bin
```

Then clone the repo to `$GOPATH/src/github.com/thingforward/slyft-cli`. That done, you can build `slyft-cli` as follows:
```
$ sudo npm install --global gulp-cli
$ npm install 
$ gulp
```

This will create a binary for your platform in the folder `bin` and a zipped archive (e.g. `dist/slyft-cli-0.1.0-darwin_1bb262da570bff653a8d8be9e785fb40.zip`) in the folder [dist](dist). You can try it by running `bin/slyft-cli` or (on Windows) `bin\slyft-cli.exe`.

What's `Gulp` doing here? It fetches any missing Go dependencies, formats and vets the source, builds the binary, and runs the tests.

You can also call `gulp build` (same as the default task), `gulp test` (just run the tests), `gulp watch` (watch source files and trigger builds when they change) individually if you prefer.

If you find that `gulp` is not recognised (or you had to skip the first step because `sudo` is not available), you can call the local copy of `gulp` installed by `npm` directly:
```
$ node node_modules/gulp/bin/gulp.js
```

### Ubuntu 

There is a Debian naming conflict where the package manager installs `nodejs` but `gulp` expects the executable to be called `node` (that being the standard name of the Node.js binary).

To solve this problem, either install `nodejs-legacy` (which adds a symlink from `/usr/bin/nodejs` to `/usr/bin/node`) or call the local gulp instance directly using `nodejs` not `node`:
```
$ nodejs node_modules/gulp/bin/gulp.js
```

### Docker

Use the `Dockerfile` to build the slyft client, use it from within a container, or copy it over to the host:

```
$ docker build -t slyft-cli .
(...)

$ docker run slyft-cli

Usage: Slyft [OPTIONS] COMMAND [arg...]
(...)

$ docker run -v $PWD:/tmpdist slyft-cli /bin/sh -c 'cp *.zip /tmpdist'
$ ls *.zip
slyft-cli-0.1.1-debian-8.6_d80891d37976c3106093b391445cec40.zip
```

### Windows
On Windows, be sure to build the program from Git Bash or a similar, unixy command prompt. `gofmt` in particular expects tools such as `diff` to be available.

When submitting pull requests, consider disabling Git's auto-detection for line endings:
```
git config --global core.autocrlf false
```
Avoiding `crlf` is important as `gofmt` standardises on Unix line endings.

## License
(C) 2016 Digital Incubation and Growth GmbH All Rights Reserved
