from starlette.routing import Route
from tabulate import tabulate


def list_routes():
    from DoboApi.application import get_app

    app = get_app()

    table = []
    for route in app.routes:

        if isinstance(route, Route):
            table.append([route.name, route.path, route.endpoint])
        else:
            table.append([route.name, '???', type(route)])

    print(tabulate(table, headers=('name', 'path', 'endpoint')))


if __name__ == "__main__":
    list_routes()
