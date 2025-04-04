"""
Test setup module for auth_api tests.

This module handles database connection setup and teardown for tests.
"""

import os

import pytest
import transaction
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker

from auth.config import Config

# Test database configuration - use the same as in config if available
TEST_DB_URL = os.environ.get('TEST_DB_URL', Config.db_url)


@pytest.fixture(scope='session')
def db_engine():
    """
    Create and return a database engine for the test session.

    This uses a single database connection for all tests.
    """
    engine = create_engine(TEST_DB_URL)
    yield engine
    engine.dispose()


@pytest.fixture(scope='session')
def db_session_factory(db_engine):
    """
    Create and return a session factory for the test database.
    """
    factory = sessionmaker(bind=db_engine)
    return factory


@pytest.fixture(scope='function')
def db_session(db_session_factory):
    """
    Create a database session wrapped in a transaction that will be rolled back.

    This ensures each test has a clean database state that doesn't affect other tests.
    """
    # Start a transaction
    with transaction.manager:
        # Get a session with transaction
        session = db_session_factory()

        # Give the session to the test
        yield session

        # Rollback the transaction after the test
        transaction.abort()


@pytest.fixture(scope='module')
def db_session_module(db_session_factory):
    """
    Create a session for an entire module that will be rolled back at the end.

    Useful when multiple tests in a module need to share database state.
    """
    # Start a transaction
    with transaction.manager:
        # Get a session with transaction
        session = db_session_factory()

        # Give the session to the test
        yield session

        # Rollback the transaction after all tests in the module
        transaction.abort()
