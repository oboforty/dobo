import enum
import os
from pathlib import Path
from tempfile import gettempdir
from datetime import timedelta

from typing import Tuple, Type, Literal

from pydantic import BaseModel
from pydantic_settings import (
    BaseSettings,
    PydanticBaseSettingsSource,
    SettingsConfigDict,
    TomlConfigSettingsSource,
)


TEMP_DIR = Path(gettempdir())
HTTPAPI_DEV = os.environ.get('HTTPAPI_DEV', 'dev')


class LogLevel(str, enum.Enum):  # noqa: WPS600
    """Possible log levels."""

    NOTSET = "NOTSET"
    DEBUG = "DEBUG"
    INFO = "INFO"
    WARNING = "WARNING"
    ERROR = "ERROR"
    FATAL = "FATAL"


class OboDBConfig(BaseModel):
    host: str = "127.0.0.1"
    port: int = 8000
    database: str


class OboDBClientConfig(BaseModel):
    tls_cert: str
    tls_key: str

    pool_size: int = 3
    metadata_cache_ttl: int = 300
    summary_cache_ttl: int = 300


class Settings(BaseSettings):
    """
    Application settings.

    These parameters can be configured
    with environment variables.
    """
    DB: OboDBConfig
    DBClient: OboDBClientConfig

    host: str = "127.0.0.1"
    port: int = 8000
    workers_count: int = 1
    reload: bool = False

    verify_email_template: str = None

    # Current environment
    environment: str = "dev"
    log_level: LogLevel = LogLevel.INFO

    # Variables for the database
    db_echo: bool = False

    @classmethod
    def settings_customise_sources(
        cls,
        settings_cls: Type[BaseSettings],
        init_settings: PydanticBaseSettingsSource,
        env_settings: PydanticBaseSettingsSource,
        dotenv_settings: PydanticBaseSettingsSource,
        file_secret_settings: PydanticBaseSettingsSource,
    ) -> Tuple[PydanticBaseSettingsSource, ...]:
        return (TomlConfigSettingsSource(settings_cls),)

    model_config = SettingsConfigDict(
        toml_file=f'settings.{HTTPAPI_DEV}.toml',
        env_file_encoding="utf-8",
    )


settings = Settings()
