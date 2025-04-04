import os
import asyncio
from datetime import datetime
from unittest import TestCase

from auth.config import Config
from auth.db import session_factory
from tests.datautils import DataUtils


def get_type_array():
    """
    This is just an array of types that will be used for testing but probably needs to move somewhere else
    """
    return ['a', 1, 1.0, 1.9999999999999999, [], {}, (), object]


bad_data_typevals_list = [
    'string',
    'unicode string',
    21,
    0,
    None,
    4.2,
    {},
    [],
    -8,
    ('', ''),
    object(),
    get_type_array,  # a function
    datetime.now(),
]

engine = None
Session = None


def run_coroutine(coroutine):
    """
    Helper function to run a coroutine synchronously.
    This allows us to test async code without using async fixtures.

    Args:
        coroutine: The coroutine to run

    Returns:
        The result of the coroutine
    """
    loop = asyncio.new_event_loop()
    try:
        return loop.run_until_complete(coroutine)
    finally:
        loop.close()


class TestBase(TestCase):
    """
    This should be the basis of all tests, with further overrides
    provided by child classes.  This should keep some basic functionality in
    mind and make test writing efficient
    """

    @classmethod
    def setUpClass(cls):
        global engine
        global Session
        cls.Config = Config
        cls.SessionFactory = session_factory
        cls.class_session = cls.SessionFactory()

    @classmethod
    def tearDownClass(cls):
        cls.class_session.rollback()

    def setUp(self):
        # We want a per-test session within the class-wide session, like a layered burrito
        self.session = self.class_session
        self.session.begin_nested()
        self.datautils = DataUtils(self.session)

    def tearDown(self):
        self.session.rollback()

    def create_user(self, extra_data=None):
        return self.datautils.create_user(extra_data)
