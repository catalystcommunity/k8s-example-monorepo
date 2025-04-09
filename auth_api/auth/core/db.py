from sqlalchemy.orm import sessionmaker, Session
from starlette.middleware.base import BaseHTTPMiddleware
from fastapi import Request
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy import (
    Engine,
    create_engine,
)
from typing import Dict, Any, List, Type, Optional, TypeVar, Union

from auth.core.config import Config
from auth.core.logger import auth_logger
from auth.core.models import Base, UserORM, SessionORM

import psycopg2

# Type variable for model conversion functions
T = TypeVar('T')


# This section is core db connection and session handling


def get_engine() -> Engine:
    return create_engine(Config.db_url)


def get_session_factory(engine) -> sessionmaker[Session]:
    factory = sessionmaker()
    factory.configure(bind=engine)
    return factory


def setupDB() -> sessionmaker[Session]:
    """
    This is the updated way to get a session factory for the DB.
    """
    new_factory = get_session_factory(get_engine())
    return new_factory


session_factory = setupDB()


def get_tm_session(use_factory):
    """
    Get a ``sqlalchemy.orm.Session`` instance backed by a transaction.

    This function will hook the session to the transaction manager which
    will take care of committing any changes.

    - When using pyramid_tm it will automatically be committed or aborted
      depending on whether an exception is raised.

    - When using scripts you should wrap the session in a manager yourself.
      For example::

          import transaction

          engine = get_engine(settings)
          session_factory = get_session_factory(engine)
          with transaction.manager:
              dbsession = get_tm_session(session_factory, transaction.manager)

    :param use_factory: the session factory we're instantiating

    """
    dbsession = use_factory()
    return dbsession


class TransactionMiddleware(BaseHTTPMiddleware):
    def __init__(self, app, logger=None):
        super().__init__(app)
        self.logger = logger
        if logger is None:
            self.logger = auth_logger

    async def dispatch(self, request: Request, call_next):
        # start session, add to request
        self.logger.info('\n\n')
        session = session_factory()
        request.state.dbsession = session
        response = await call_next(request)
        if response.status_code >= 200 and response.status_code <= 299:
            self.logger.info('COMMIT')
            session.commit()
        else:
            self.logger.info('ROLLBACK')
            session.rollback()
        session.close()
        return response


# This section contains db conversion utilities for use in view handling


def orm_to_dict(
    orm_model: Base, exclude_fields: Optional[List[str]] = None
) -> Dict[str, Any]:
    """
    Convert an SQLAlchemy ORM model to a dictionary

    Args:
        orm_model: SQLAlchemy ORM model instance
        exclude_fields: Optional list of field names to exclude

    Returns:
        Dictionary representation of the model
    """
    if exclude_fields is None:
        exclude_fields = []

    # Special handling for password and salt fields - always exclude them
    if hasattr(orm_model, 'password') and 'password' not in exclude_fields:
        exclude_fields.append('password')
    if hasattr(orm_model, 'salt') and 'salt' not in exclude_fields:
        exclude_fields.append('salt')

    result = {}
    for column in orm_model.__table__.columns:
        if column.name not in exclude_fields:
            result[column.name] = getattr(orm_model, column.name)

    return result


def orm_to_pydantic(orm_model: Base, pydantic_model: Type[T]) -> T:
    """
    Convert an SQLAlchemy ORM model to a Pydantic model

    Args:
        orm_model: SQLAlchemy ORM model instance
        pydantic_model: Pydantic model class

    Returns:
        Instance of the Pydantic model
    """
    # Convert ORM model to dict first, then to Pydantic model
    model_dict = orm_to_dict(orm_model)
    return pydantic_model.model_validate(model_dict)


def dict_to_orm(model_dict: Dict[str, Any], orm_class: Type[Base]) -> Base:
    """
    Convert a dictionary to an SQLAlchemy ORM model

    Args:
        model_dict: Dictionary containing model data
        orm_class: SQLAlchemy ORM model class

    Returns:
        Instance of the ORM model
    """
    orm_instance = orm_class()

    # Copy values from dict to ORM instance
    for column in orm_instance.__table__.columns:
        if column.name in model_dict:
            setattr(orm_instance, column.name, model_dict[column.name])

    return orm_instance


def make_set_of_field_names(field_names=None):
    """
    Utility to normalize field names

    Args:
        field_names: Field names as string, list, or attributes with keys

    Returns:
        List of field name strings
    """
    if field_names:
        field_names = (
            [field_names] if not isinstance(field_names, list) else field_names
        )
        if not all(isinstance(x, str) for x in field_names):
            # We don't care if it's a Column or an InstrumentedAttribute or whatever, so long as we get a key from it
            if not all(hasattr(x, 'key') for x in field_names):
                raise ValueError('not all fields are strings or have names')
            # Make everything a string instead of a column, just to make things simpler
            field_names = [x.key if hasattr(x, 'key') else x for x in field_names]
        return field_names
    else:
        return []


def dict_from_row(row, remove_fields=None, sub_values=None):
    """
    The web framework is not aware of the model classes used for our database structures, and it should not be, so
    this function translates between an arbitrary class and a dictionary of field: value pairs for passing
    to a jinja template, typically for rendering to JSON rather than HTML.  It should not always be necessary
    unless the class over that area of responsibility is not built expressly for web framework consumption.
    :param row: A class, typically returned as a row from a query, but not a "ResultRow" object
    :param remove_fields: A list of fields, by string or sqlalchemy object attribute, to remove from the returned dict
    :param sub_values: A list of fields to be further loaded into the return dict, only if present as column names
    :return: A dictionary representation of all the non-private attributes of the row given
    """
    # For compatibility with new code, use orm_to_dict if possible
    if hasattr(row, '__table__'):
        # Convert remove_fields to a list of strings
        exclude = make_set_of_field_names(remove_fields) if remove_fields else []
        return orm_to_dict(row, exclude)

    # Fall back to old implementation for non-SQLAlchemy objects
    retdict = {}
    remove_fields = make_set_of_field_names(remove_fields)
    sub_values = make_set_of_field_names(sub_values)

    for public_key in [
        i.name for i in row.__table__.columns if i.name not in remove_fields
    ]:
        value = getattr(row, public_key)
        if value is not None:
            retdict[public_key] = value
        else:
            retdict[public_key] = None
    for key in [i for i in sub_values if hasattr(row, i)]:
        sub = getattr(row, key)
        # This used to be the way it was, but Tod didn't know why it would be a list but also a dict
        # if isinstance(sub, list) and not isinstance(sub, dict):
        if isinstance(sub, list):
            retdict[key] = [
                dict_from_row(i, remove_fields=remove_fields)
                for i in sub
                if i not in remove_fields
            ]
        else:
            retdict[key] = sub._to_dict()
    return retdict


def array_of_dicts_from_map(map, remove_fields=None, sub_values=None):
    """
    Helper function for processing an entire resultset from some sqlalchemy object queries, assuming
    :param map: The resultset
    :param remove_fields: A list of fields, by string or sqlalchemy object attribute, to remove from the returned dict
    :param sub_values: A list of fields, by string or sqlalchemy object attribute, to get sub_values for on each object
    :return: An array of dictionaries as rendered one at a time from the function appropriate
    """
    # It's faster to do this once ahead of the dataset instead of redoing it every call
    remove_fields = make_set_of_field_names(remove_fields)
    sub_values = make_set_of_field_names(sub_values)
    return [
        x
        if not isinstance(x, declarative_base())
        else dict_from_row(x, remove_fields, sub_values)
        for x in map.values()
    ]


def array_of_dicts_from_array_of_models(models, remove_fields=None, sub_values=None):
    """
    Helper function for processing an entire resultset from some sqlalchemy object queries
    :param models: The resultset
    :param remove_fields: A list of fields, by string or sqlalchemy object attribute, to remove from the returned dict
    :param sub_values: A list of fields, by string or sqlalchemy object attribute, to get sub_values for on each object
    :return: An array of dictionaries as rendered one at a time from the function appropriate
    """
    # It's faster to do this once ahead of the dataset instead of redoing it every call
    remove_fields = make_set_of_field_names(remove_fields)
    sub_values = make_set_of_field_names(sub_values)
    return [dict_from_row(x, remove_fields, sub_values) for x in models]


# I'd really like to never have to use this ever again, but just in case, we're leaving the code

# def columnar_object_from_array_of_models(models, remove_fields=None, sub_values=None, renames=None):
#     """
#     Helper function for processing an entire resultset from some sqlalchemy object queries
#     :param models: The resultset
#     :param remove_fields: A list of fields, by string or sqlalchemy object attribute, to remove from the returned dict
#     :param sub_values: A list of fields, by string or sqlalchemy object attribute, to get sub_values for on each object
#     :param renames: A dict of fields to rename from keys to values
#     :return: A dictionary with an array of values as rendered one at a time from the function appropriate
#     """
#     # It's faster to do this once ahead of the dataset instead of redoing it every call
#     remove_fields = make_set_of_field_names(remove_fields)
#     sub_values = make_set_of_field_names(sub_values)
#     result={}
#     for item in models:
#         for col in [x.name for x in item.__table__.columns if x.name not in remove_fields and not x.name.startswith('_')]:
#             newval = col
#             if col in renames:
#                 newval = renames[col]
#             try:
#                 result[newval].append(getattr(item, col))
#             except KeyError:
#                 result[newval]=[getattr(item, col)]
#     return result


def array_of_dicts_from_array_of_keyed_tuples(keyed_tuples):
    """
    Helper function for processing an entire resultset of a query that touched multiple tables,
    resulting in a list of keyed tuples
    :param keyed_tuples: The resultset, as a list of keyed tuples
    :return: An array of dictionaries as rendered one at a time from the function appropriate
    """
    return [x._asdict() for x in keyed_tuples]


def sqlobj_from_dict(obj, values):
    """
    Merge in items in the values dict into our object if it's one of our columns.
    """
    for c in obj.__table__.columns:
        if c.name in values:
            setattr(obj, c.name, values[c.name])
    # This return isn't necessary as the obj referenced is modified, but it makes it more
    # complete and consistent
    return obj
