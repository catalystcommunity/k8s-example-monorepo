import os
import uuid
from random import randint
from typing import Any

from sqlalchemy.orm import Session as SQLSession

from auth.core.db import sqlobj_from_dict
from auth.core.models import (
    SessionORM,
    UserORM,
    UserRole,
)
from auth.core.passwords import hash_password

# Alias for backward compatibility
User = UserORM


def random_with_n_digits(n: int) -> int:
    """Generate a random number with n digits"""
    range_start = 10 ** (n - 1)
    range_end = (10**n) - 1
    return randint(range_start, range_end)


class DataUtils:
    """
    This class handles creation of database objects for tests.

    It encapsulates the boilerplate of creating database records with sensible defaults
    while allowing test-specific overrides. This makes tests more readable by focusing
    on the test logic rather than data setup.
    """

    def __init__(self, session: SQLSession):
        """
        Initialize with a database session.

        Args:
            session: An SQLAlchemy session for database operations
        """
        self.session = session

    def create_user(
        self, data: dict[str, Any] | None = None, return_object: bool = True
    ) -> UserORM | uuid.UUID:
        """
        Create a user with default values that can be overridden.

        Args:
            data: Dictionary of attribute overrides
            return_object: Whether to return the object or just the ID

        Returns:
            Either the User object or its ID
        """
        u = UserORM()

        if data is None:
            data = {}

        # Apply any provided values
        sqlobj_from_dict(u, data)

        # Set default values for required fields if not provided
        if u.username is None:
            u.username = f'testuser_{str(uuid.uuid4())[:8]}'
        else:
            u.username = u.username.lower()

        if u.email is None:
            u.email = f'{u.username}@example.com'
        else:
            u.email = u.email.lower()

        # Generate random salt if not provided
        if u.salt is None:
            u.salt = os.urandom(256)

        # Hash password with PBKDF2 if provided as string
        if u.password is None:
            password_str = f'password_{str(uuid.uuid4())[:8]}'
            u.password = hash_password(password_str, u.salt)
        elif isinstance(u.password, str):
            u.password = hash_password(u.password, u.salt)

        # Set default role if not provided
        if u.roles is None or len(u.roles) == 0:
            u.roles = [UserRole.user]

        # Save to database
        self.session.add(u)
        self.session.flush()
        self.session.refresh(u)

        if return_object:
            return u
        return u.user_id

    def create_session(
        self, data: dict[str, Any] | None = None, return_object: bool = True
    ) -> SessionORM | uuid.UUID:
        """
        Create a session with default values that can be overridden.

        Args:
            data: Dictionary of attribute overrides
            return_object: Whether to return the object or just the ID

        Returns:
            Either the Session object or its ID
        """
        s = SessionORM()
        if data is None:
            data = {}

        # Apply any provided values
        sqlobj_from_dict(s, data)

        # Create user if not provided
        if s.user_id is None:
            user_data = {}
            if 'user' in data:
                user_data = data['user']
            user = self.create_user(user_data)
            s.user_id = user.user_id

        # Generate token if not provided
        if s.token is None:
            s.token = str(uuid.uuid4())

        # Save to database
        self.session.add(s)
        self.session.flush()
        self.session.refresh(s)

        if return_object:
            return s
        return s.session_id
