import os
import re
from datetime import datetime
from uuid import uuid4

from fastapi import Body, APIRouter, Request, Response
from email_validator import validate_email, EmailNotValidError
from sqlalchemy.orm import Session as SQLSession

from auth.core.models import SessionORM, UserORM, UserRole, Session, UserResponse
from auth.core.passwords import hash_password
from auth.core.error_dict import error_dict

# Sphinx doc stuff
from auth.core.db import dict_from_row, orm_to_pydantic

# Alias for backward compatibility
User = UserORM

usersRouter = APIRouter(prefix='/api', tags=['users'])
removals = ['password', 'salt']
recovery_template = """
Hi %s,

Your account password has been reset in response to a valid forgotten password request.
Please use the following temporary password to log in, and then change your password immediately:

%s

For any questions, don't hesitate to contact us at The Website

- The Website Team
We know how to web!
"""


def username_in_use(username, dbsession):
    """
    Returns true if a username exists
    :param username: a string username
    :param dbsession: a db session
    :return: True if a username exists, otherwise False
    """
    # Check if we have a valid session
    if not isinstance(dbsession, SQLSession):
        return False

    # Use classic SQLAlchemy query approach for compatibility
    user = dbsession.query(UserORM).filter_by(username=username.lower()).first()
    return user is not None


@usersRouter.post('/users')
async def users_post_view(request: Request, response: Response):
    json_body = await request.json()
    username = json_body.get('username')
    if not isinstance(username, str):
        response.status_code = 400
        return {
            'd': error_dict(
                'api_errors',
                'username, email, and password are all required string fields',
            )
        }
    if username_in_use(json_body['username'], request.state.dbsession):
        response.status_code = 400
        return {
            'd': error_dict(
                'verification_error',
                'username already in use: %s' % json_body['username'],
            )
        }

    requires = ['email', 'password']
    if not all(field in json_body for field in requires) or not all(
        isinstance(json_body.get(field), str) for field in json_body
    ):
        response.status_code = 400
        return {
            'd': error_dict(
                'api_errors',
                'username, email, and password are all required string fields',
            )
        }

    user = UserORM()
    user.salt = os.urandom(256)
    user.password = hash_password(json_body['password'], user.salt)
    user.username = json_body['username'].lower()
    user.email = json_body['email'].lower()
    user.roles = [UserRole.user]

    request.state.dbsession.add(user)
    request.state.dbsession.flush()
    request.state.dbsession.refresh(user)

    s = SessionORM()
    s.user_id = user.user_id
    s.token = str(uuid4())
    request.state.dbsession.add(s)
    request.state.dbsession.flush()
    request.state.dbsession.refresh(s)

    # Create response using the data access functions
    result = dict_from_row(user, remove_fields=removals)
    result['session'] = dict_from_row(s, remove_fields=removals)

    return {'d': result}


@usersRouter.put('/users')
async def users_put_view(request: Request, response: Response):
    """
    Used for forgotten password requests
    """
    # TODO Make tests for this
    # TODO Make this a temporary reset link thing instead of actually resetting passwords to random
    if request.state.user is not None:
        response.status_code = 400
        return {'d': error_dict('api_errors', 'you are already logged in, ignoring')}
    username = request.json_body.get('username')
    email = request.json_body.get('email')
    if not isinstance(username, str) or not isinstance(email, str):
        response.status_code = 400
        return {
            'd': error_dict(
                'api_errors', 'username and email are required string fields'
            )
        }
    user = (
        request.state.dbsession.query(User)
        .filter(User.username == username.lower())
        .one_or_none()
    )
    if user is None or user.email != email.lower():
        response.status_code = 400
        return {'d': error_dict('api_errors', 'invalid state found')}
    uid = uuid4()
    newpass = uid.hex
    user.salt = os.urandom(256)
    user.password = hash_password(newpass, user.salt)

    # TODO send an email when a recovery happens using the template above
    return {'d': 'recovery email sent'}


@usersRouter.get('/users/{user_id}')
async def user_id_get_view(user_id: int, request: Request, response: Response):
    if request.state.user is None:
        response.status_code = 400
        return {'d': error_dict('api_errors', 'not authenticated for this request')}
    if not user_id or user_id != request.state.user.id:
        response.status_code = 400
        return {'d': error_dict('api_errors', 'not authenticated for this request')}
    user = request.state.user
    result = dict_from_row(user, remove_fields=removals)
    return {'d': result}


@usersRouter.put('/users/{user_id}')
async def user_id_put_view(request: Request, response: Response):
    if request.state.user is None:
        response.status_code = 400
        return {'d': error_dict('api_errors', 'not authenticated for this request')}
    if (
        not request.matchdict.get('user_id')
        or int(request.matchdict.get('user_id')) != request.state.user.user_id
    ):
        response.status_code = 400
        return {'d': error_dict('api_errors', 'not authenticated for this request')}
    valid_types = {
        'email': str,
        'pin': str,
        'timezone': datetime,
        'infoemails': bool,
    }
    email = request.json_body.get('email')
    if not isinstance(email, str):
        response.status_code = 400
        return {'d': error_dict('api_errors', 'email invalid: must be a string')}
    try:
        v = validate_email(email)  # validate and get info
        email = v.normalized  # replace with normalized form
    except EmailNotValidError as e:
        # email is not valid, exception message is human-readable
        response.status_code = 400
        return {'d': error_dict('api_errors', 'email invalid: %s' % e)}
    password = request.json_body.get('password')
    # Password must be optional, since they don't know the old value
    if password is not None:
        if not isinstance(password, str):
            response.status_code = 400
            return {'d': error_dict('api_errors', 'password must be a string')}
        if len(password) < 8:
            response.status_code = 400
            return {
                'd': error_dict('api_errors', 'password must be at least 8 characters')
            }
        request.state.user.password = hash_password(password, request.state.user.salt)

    request.state.user.email = email

    request.state.dbsession.flush()
    request.state.dbsession.refresh(request.state.user)
    return {'d': dict_from_row(request.state.user, remove_fields=removals)}
