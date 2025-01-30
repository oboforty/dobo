# DoboApi

## Set up

### Codebase

    poetry install
    poetry run python -m DoboApi

You can find swagger documentation at `/api/docs`.

If you want to migrate your database, you should run following commands:

### Migrations

    alembic upgrade "head"

    alembic revision --autogenerate

    # Revert everything.
     alembic downgrade base


### Running tests
Without postgres DB it's as easy as

    pytest -vv .


---

## Deploy

### Manual Deploy

    poetry export --without-hashes --format=requirements.txt > requirements.txt


Build the multi-stage Dockerfile
    docker build -f deploy.Dockerfile .

On remote:
    
    sudo mkdir /app

---

## Codebase
