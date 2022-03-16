FROM python:3.9-slim
RUN pip3 install pipenv

WORKDIR /app
ADD Pipfile Pipfile.lock /app/
RUN pipenv install --system --deploy
RUN mkdir /app/site

ENTRYPOINT [ "mkdocs" ]
CMD ["build"]
