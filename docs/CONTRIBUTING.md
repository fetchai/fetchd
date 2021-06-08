# Documentation Setup

## Prerequisites

Make sure that you have pipenv installed on your system:

    pip3 install pipenv

Once installed navigate to this folder in the project

    cd fetchd/docs

Make sure all the dependencies are installed

    pipenv install -d

## Updating the docs

Once the dependencies are setup you must activate the environment with the following commands

    pipenv shell

This step should update your terminal prompt and you will be able to see that the command `mkdocs` is installed in your path


    which mkdocs

Finally, to start the development server run the following command


    mkdocs serve

This will listen for changes on the filesystem and automatically update the documentation.
