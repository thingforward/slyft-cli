# slyft-cli
command line client to slyft-server

## Run slyft-cli
To run `slyft-cli`, fetch the appropriate zip file from the folder [dist](dist), extract and run:
```
$ ./slyft-cli
```

## Build slyft-cli
Before you begin, make sure you have Golang and Node.js installed, and clone the repo to `$GOPATH/src/github.com/thingforward/slyft-cli`. That done, you can build `slyft-cli` as follows:
```
$ sudo npm install --global gulp-cli
$ npm install 
$ gulp
```

What's `Gulp` doing here? It fetches any missing Go dependencies, formats and vets the source, builds the binary, and runs the tests.

You can also call `gulp build` (same as the default task), `gulp test` (just run the tests), `gulp watch` (watch source files and trigger builds when they change) individually if you prefer.

# License
(C) 2016 Digital Incubation and Growth GmbH All Rights Reserved
