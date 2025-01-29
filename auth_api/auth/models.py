import enum
import uuid
from datetime import datetime
from typing import List

from sqlalchemy import TIMESTAMP, Column, FetchedValue, text
from sqlalchemy.types import LargeBinary
from sqlmodel import ARRAY, Field, SQLModel, Enum

"""
Typing information is for python validation, not for DB operations. DB information is in migrations in SQL.
"""


class UserRole(str, enum.Enum):
    user = 'user'
    support = 'support'
    admin = 'admin'


class User(SQLModel):
    """
    What it says on the tin, a User represents the metadata for someone to have an account and login
    """

    __tablename__: str = 'users'

    id: uuid.UUID | None = Field(sa_column_kwargs={'server_default': 'generate_ulid()'})
    created_at: datetime | None = Field(
        sa_column=Column(
            TIMESTAMP(timezone=True),
            nullable=False,
            # sa_column_kwargs={'server_default':"timezone('utc', now())"},
            server_default="timezone('utc', now())",
        )
    )
    updated_at: datetime | None = Field(
        sa_column=Column(
            TIMESTAMP(timezone=True),
            nullable=False,
            server_default="timezone('utc', now())",
            server_onupdate=FetchedValue(),
        )
    )
    username: str
    email: str
    password: bytes = Field(sa_column=Column(LargeBinary))
    salt: bytes = Field(sa_column=Column(LargeBinary))
    roles: List[UserRole] = Field(
        sa_column=Column(ARRAY(Enum(UserRole))), default=[UserRole.user]
    )


class Session(SQLModel):
    """
    Sessions are the logins of a particular user, originally limited to one per user because
    John didn't see a reason to support anything else. This is a pattern.
    """

    __tablename__: str = 'sessions'

    id: uuid.UUID | None = Field(sa_column_kwargs={'server_default': 'generate_ulid()'})
    user_id: uuid.UUID | None = Field(
        sa_column_kwargs={'server_default': 'generate_ulid()'}, foreign_key='users.id'
    )
    created_at: datetime | None = Field(
        sa_column=Column(
            TIMESTAMP(timezone=True),
            nullable=False,
            server_default="timezone('utc', now())",
        )
    )
    updated_at: datetime | None = Field(
        sa_column=Column(
            TIMESTAMP(timezone=True),
            nullable=False,
            server_default="timezone('utc', now())",
            server_onupdate=FetchedValue(),
        )
    )
    token: str
