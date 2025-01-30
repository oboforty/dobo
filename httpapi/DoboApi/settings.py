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
HELLAS_ENV = os.environ.get('HELLAS_ENV', 'dev')


class LogLevel(str, enum.Enum):  # noqa: WPS600
    """Possible log levels."""

    NOTSET = "NOTSET"
    DEBUG = "DEBUG"
    INFO = "INFO"
    WARNING = "WARNING"
    ERROR = "ERROR"
    FATAL = "FATAL"


class DBConfig(BaseModel):
    dsn: str
    database: str
    test_database: str = 'test_db'

    @property
    def db_url(self) -> str:
        """
        Assemble database URL from settings.

        :return: database URL.
        """
        return self.dsn.format(database=self.database)


class OAuthConfig(BaseModel):
    algorithm: str
    issuer: str
    audience: str


class Settings(BaseSettings):
    """
    Application settings.

    These parameters can be configured
    with environment variables.
    """
    DB: DBConfig
    OAuth: OAuthConfig

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
        toml_file=f'settings.{HELLAS_ENV}.toml',
        env_file_encoding="utf-8",
    )


settings = Settings()
