# Building the Ledger

## Prerequisites

- Go 1.16+ (installation instructions available [here](https://golang.org/dl/))
- Packages: `make`, `gcc` (on ubunty, install them with `sudo apt-get update && sudo apt-get install -y make gcc`). For mac ist is recomended to use the clang compiler that comes with xcode. No additional installation is needed.

## Building the code

Download the latest released version from github and build the project using the following commands:

```
gh repo clone fetchai/fetchd && cd fetchd
```

Then select the particular tag version you want to build, ie:

```bash
git switch v0.8.7
```

Then build the code with the command:

```bash
make build
```

This will generate the `./build/fetchd` binary.

For non-developer users we recommend that the user installs the binaries into their system. This can be done with the following command:

```bash
make install
```

This will install the binaries in the directory specified by your `$GOBIN` environment variable (default to `~/go/bin`).


```bash
which fetchd
```

This should return a path such as `~/go/bin/fetchd` (might be different depending on your actual go installation).

> If you get no output, or an error such as `which: no fetchd in ...`, possible cause can either be that `make install` failed with some errors or that your go binary folder (default: ~/go/bin) is not in your `PATH`.
>
> To add the ~/go/bin folder to your PATH, add this line at the end of your ~/.bashrc:
>```
>export PATH=$PATH:~/go/bin
>```
>
>and reload it with:
>
>```
>source ~/.bashrc
>```

You can also verify that you are running the correct version 

```bash
fetchd version
```

This should print a version number that must be compatible with the network you're connecting to (see the [network page](../networks/) for the list of supported versions per network).

## FAQ

- **Error: failed to parse log level (main:info,state:info,:error): Unknown Level String: 'main:info,state:info,:error', defaulting to NoLevel**

This means you had a  pre-stargate version of fetchd (<= v0.7.x), and just installed a stargate version (>= v0.8.x), you'll need to remove the previous configuration files with:

```bash
rm ~/.fetchd/config/app.toml ~/.fetchd/config/config.toml
```
