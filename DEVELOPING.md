# Development Guidelines

- [Getting the Source](#get)
- [Setting up a New Development Environment](#setup)
- [Development](#dev)
- [Testing](#test)
- [Contributing](#contributing)

## <a name="get"></a> Getting the Source

<!-- markdown-link-check-disable -->
1. Fork the [repository](https://github.com/fetchai/fetchd.git).
2. Clone your fork of the repository:
    <!-- markdown-link-check-enable -->

   ``` shell
   git clone https://github.com/fetchai/fetchd.git
   ```

3. Define an `upstream` remote pointing back to the main FetchD repository:

   ``` shell
   git remote add upstream https://github.com/fetchai/fetchd.git
   ```

## <a name="setup"></a> Setting up a New Development Environment

The easiest way to get set up for development is to install Python (`3.9` to `3.12`) and [poetry](https://pypi.org/project/poetry/), and then run the following from the top-level project directory:

```bash
  cd python
  poetry install
  poetry shell
  pre-commit install
```

## <a name="dev"></a>Development

When developing for `fetchd` make sure to have the poetry shell active. This ensures that linting and formatting will automatically be checked during `git commit`.

We are using [Ruff](https://github.com/astral-sh/ruff) with added rules for formatting and linting.
Please consider adding `ruff` to your IDE to speed up the development process and ensure you only commit clean code.

Alternately you can invoke ruff by typing the following from within the `./python` folder

```bash
  ruff check --fix && ruff format
```

## <a name="test"></a>Testing

To run tests use the following command:

```bash
  pytest
```

## <a name="contributing"></a>Contributing

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
