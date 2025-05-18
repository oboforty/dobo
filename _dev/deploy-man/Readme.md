# Install manually

Copy:

- ./nginx
- fastapi.service to `/etc/systemd/system/fastapi.service`
- app files to `/app/HellasApi`
- jwks private file to `/app/jwks/`

Envvars:

    DAVE_ENV=prod

Run:
    pip install -r requirements.txt
    pip install uwsgi

    sudo systemctl daemon-reload
    sudo systemctl enable fastapi
    sudo systemctl start fastapi

    sudo systemctl restart nginx
