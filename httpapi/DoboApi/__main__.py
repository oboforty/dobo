import uvicorn

from DoboApi.settings import settings


def main() -> None:
    """Entrypoint of the application."""
    uvicorn.run(
        "DoboApi.application:get_app",
        host=settings.host,
        port=settings.port,
        reload=settings.reload,
        log_level=settings.log_level.value.lower(),
        reload_dirs=["DoboApi"],
        # factory=True,
    )


if __name__ == "__main__":
    main()
