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


This will create a binary for your platform in the folder `bin` and a zip archive (e.g. `slyft-cli-0.1.0-darwin.zip`) in the folder [dist](dist). You can try it by running `bin/slyft-cli` or (on Windows) `bin\slyft-cli.exe`.

What's `Gulp` doing here? It fetches any missing Go dependencies, formats and vets the source, builds the binary, and runs the tests.

You can also call `gulp build` (same as the default task), `gulp test` (just run the tests), `gulp watch` (watch source files and trigger builds when they change) individually if you prefer.

Note for Ubuntu users: due to a naming conflict where `nodejs` is installed but the package expects `node` to be present, it's quickest to use `nodejs node_modules/gulp/bin/gulp.js` instead of `gulp`.

# License
(C) 2016 Digital Incubation and Growth GmbH All Rights Reserved
