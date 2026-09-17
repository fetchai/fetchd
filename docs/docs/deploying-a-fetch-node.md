# Deploying a Fetch.AI node

Supported platforms

* MacOS Darwin 10.13x and higher (64bit)
* Ubuntu 18.04 (x86_64)

We plan to support all major platforms in the future.

## Requirements
Before you can get going, make sure you’ve got the following requirements installed.

If you’re on Mac, you’ll need to install openssl:

`brew install cmake openssl`

If you’re on Ubuntu:

`sudo apt-get install libssl-dev cmake python3-dev clang`

With those requirements fulfilled, download our ledger from our release branch. You are also going to need to download the following submodules:

`git clone https://github.com/fetchai/ledger.git`
`cd ledger`
`mkdir build`
`git submodule update --init --recursive`

You might want to checkout the branch `v0.3.1-rc1` you may need to `git fetch` to see it, but this is our latest stable.

## Building the code
The project uses cmake so follow this build procedure:

`cd build`
`cmake ../`
`make -j`

## Running a single node locally
Open terminal and go to the build directory within the ledger directory.

`./apps/constellation/constellation -block-interval 3000 -standalone -port 8000`

## Running a private network
Open terminal and go to the build directory within the ledger directory.

In one terminal window:

`./apps/constellation/constellation -block-interval 3000 -private-network -port 8020 -peers tcp://127.0.0.1:8001`

In another terminal window:

`./apps/constellation/constellation -block-interval 3000 -private-network -port 8040 -peers tcp://127.0.0.1:8021`

## Further reading
We cover installation in further detail on our community site.

With this information you should be able to deploy a node on the Fetch.AI network. If you need help, email us at support@fetch.ai. You can also join our telegram channel.

## License
Fetch Ledger is licensed under the Apache software license (see LICENSE file). Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an \"AS IS\" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either \ express or implied.

Fetch.AI makes no representation or guarantee that this software (including any third-party libraries) will perform as intended or will be free of errors, bugs or faulty code. The software may fail which could completely or partially limit functionality or compromise computer systems. If you use or implement the ledger, you do so at your own risk. In no event will Fetch.AI be liable to any party for any damages whatsoever, even if it had been advised of the possibility of damage.

As such this codebase should be treated as experimental and does not contain all currently developed features. Fetch.AI will be delivering regular updates.
