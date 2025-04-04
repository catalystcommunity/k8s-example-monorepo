import enum
import uuid
import hashlib
import binascii
from datetime import datetime
from typing import List, Optional

from sqlalchemy import Column, TIMESTAMP, ARRAY, String, LargeBinary, Boolean, func
from sqlalchemy.dialects.postgresql import UUID, ENUM
from sqlalchemy.orm import declarative_base
from pydantic import BaseModel, ConfigDict, Field

"""
Typing information is for python validation, not for DB operations. 
DB information is in migrations in SQL.

We separate SQLAlchemy ORM models from Pydantic models for better compatibility.
"""

# SQLAlchemy Base for ORM models
Base = declarative_base()


# Enum for user roles
class UserRole(str, enum.Enum):
    user = 'user'
    support = 'support'
    admin = 'admin'


# SQLAlchemy ORM models
class UserORM(Base):
    """
    SQLAlchemy ORM model for users
    """

    __tablename__ = 'users'

    user_id = Column(UUID, primary_key=True, server_default='generate_ulid()')
    created_at = Column(
        TIMESTAMP(timezone=True),
        nullable=False,
        server_default="timezone('utc', now())",
    )
    updated_at = Column(
        TIMESTAMP(timezone=True),
        nullable=False,
        server_default="timezone('utc', now())",
        onupdate=func.now(),
    )
    username = Column(String, nullable=False, unique=True)
    email = Column(String, nullable=False)
    password = Column(LargeBinary, nullable=False)
    salt = Column(LargeBinary, nullable=False)
    roles = Column(
        ARRAY(ENUM(UserRole, name='user_role')), nullable=False, default=[UserRole.user]
    )

    def __repr__(self):
        return f'<User(id={self.user_id}, username={self.username})>'


class SessionORM(Base):
    """
    SQLAlchemy ORM model for sessions
    """

    __tablename__ = 'sessions'

    session_id = Column(UUID, primary_key=True, server_default='generate_ulid()')
    user_id = Column(UUID, nullable=False)
    created_at = Column(
        TIMESTAMP(timezone=True),
        nullable=False,
        server_default="timezone('utc', now())",
    )
    updated_at = Column(
        TIMESTAMP(timezone=True),
        nullable=False,
        server_default="timezone('utc', now())",
        onupdate=func.now(),
    )
    token = Column(String, nullable=False)
    rotated = Column(Boolean, nullable=False, default=False)

    def __repr__(self):
        return f'<Session(id={self.session_id}, user_id={self.user_id})>'


# Pydantic models for API validation and responses
class UserBase(BaseModel):
    """Base User model with common attributes"""

    username: str
    email: str
    roles: List[UserRole] = [UserRole.user]

    model_config = ConfigDict(
        from_attributes=True,  # Allow converting ORM models to Pydantic models
    )


class UserCreate(UserBase):
    """User creation model with password"""

    password: str


class UserResponse(UserBase):
    """User response model with ID and timestamps"""

    user_id: uuid.UUID
    created_at: datetime
    updated_at: datetime


class Session(BaseModel):
    """Session model for API responses"""

    session_id: uuid.UUID
    user_id: uuid.UUID
    created_at: datetime
    updated_at: datetime
    token: str
    rotated: bool = False
    verifier: Optional[str] = None

    model_config = ConfigDict(
        from_attributes=True,  # Allow converting ORM models to Pydantic models
    )


# Type aliases for backwards compatibility
# These should make transitioning from SQLModel easier
User = UserORM
# Use Session from the Pydantic models for API responses
