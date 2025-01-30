rm hellas.db

alembic upgrade "head"
echo "$PWD"

python scripts/add_fixtures.py

read -p "Press enter to continue"
