import hashlib
from copy import deepcopy
from uuid import uuid4
from random import randint

from datetime import datetime
from sqlalchemy import func
from sqlalchemy.orm import Session

from auth.models import (
    User,
    Session,
)
from auth.db import sqlobj_from_dict


def random_with_n_digits(n):
    range_start = 10 ** (n - 1)
    range_end = (10**n) - 1
    return randint(range_start, range_end)


class DataUtils(object):
    """
    This class will handle creation of specific data types for default
    records and will be reusable with a variety of tests and scenarios for
    testing environment setup.  The goal is to make code more readable in a
    test and get the data boilerplate out of the testing files for a given module,
    so we can just focus on writing tests for that module, regardless of what
    it might be.

    As an example, if one needs an object X, rather than defining methods of
    creating the database objects, we can simply make the data in the test, and
    let the redundant coding bits of dealing with the database be a function here.
    """

    def __init__(self, session: Session):
        """
        Creation of basic state to handle all our data operations.
        """
        self.session = session

    def create_user(self, spec_data=None, return_object=True):
        """
        Make a customer object, return the actual object with spec_data overriding values for further manipulation unless set to false.
        :param spec_data: A dictionary containing the data keyed on db model object attribute
        :param return_object: Whether to return the object or not, defaulting to True
        :return: a customer db model
        """
        u = User()

        if spec_data is None:
            spec_data = {}
        sqlobj_from_dict(u, spec_data)

        if u.id is None:
            u.id = self.session.query(func.nextval('users_id_seq')).scalar()
        if u.username is None:
            u.username = 'generated%d' % u.id
        if u.email is None:
            u.email = 'Test%d@example.com' % u.id
        if u.salt is None:
            u.salt = 'generated_salt%d' % u.id
        if u.password is None:
            u.password = 'generated_pass%d' % u.id
        if u.infoemails is None:
            u.infoemails = True

        if isinstance(u.salt, str):
            s = hashlib.sha512()
            s.update(u.salt.encode('utf-8'))
            u.salt = s.digest()
        if isinstance(u.password, str):
            m = hashlib.sha512()
            m.update(u.password.encode('utf-8'))
            m.update(u.salt)
            u.password = m.digest()

        self.session.add(u)
        self.session.flush()
        self.session.refresh(u)
        if return_object:
            return u
        return u.id

    def create_session(self, spec_data=None, return_object=True):
        """
        Make a session object, return the actual object with spec_data overriding values for further manipulation unless set to false.
        :param spec_data: A dictionary containing the data keyed on db model object attribute
        :param return_object: Whether to return the object or not, defaulting to True
        :return: a session db model
        """
        s = Session()
        if spec_data is None:
            spec_data = {}
        sqlobj_from_dict(s, spec_data)

        if s.user_id is None:
            s.user_id = self.create_user(spec_data).id
        if s.token is None:
            s.token = uuid4()

        self.session.add(s)
        self.session.flush()
        self.session.refresh(s)
        if return_object:
            return s
        return s.id
