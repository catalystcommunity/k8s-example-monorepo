import binascii
from datetime import datetime, timedelta
from uuid import uuid4
from fastapi import APIRouter, Request, Response, status

from auth.models import SessionORM, UserORM
from auth.error_dict import error_dict
from auth.passwords import hash_password
from auth.verification import generate_verifier

# Sphinx doc stuff
from auth.db import dict_from_row

sessions_desc = """
Work with sessions for user accounts
"""
# sessions_svc = Service(name='sessions', path='/api/sessions', description=sessions_desc, renderer='json')
sessionsRouter = APIRouter(prefix='/api', tags=['sessions'])


@sessionsRouter.post('/sessions', summary='new session')
async def sessions_post_view(request: Request, response: Response):
    """
    This will begin a new session given a username and password
    """
    json_body = await request.json()
    if request.state.user is not None and json_body.get('token') is not None:
        # Our request validated their token, so just get that token
        session = (
            request.state.dbsession.query(SessionORM)
            .filter(SessionORM.token == json_body['token'])
            .one()
        )
        result = dict_from_row(session)
        # Add verifier
        result['verifier'] = generate_verifier(
            str(request.state.user.user_id), request.state.user.salt, session.token
        )
        return {'d': result}
    username = json_body.get('username')
    if username is None or not isinstance(username, str):
        response.status_code = 400
        return {'d': error_dict('api_errors', 'no valid username provided')}
    password = json_body.get('password')
    if password is None or not isinstance(password, str):
        response.status_code = 400
        return {'d': error_dict('api_errors', 'no valid password provided')}
    user = (
        request.state.dbsession.query(UserORM)
        .filter(UserORM.username == username.lower())
        .one_or_none()
    )
    if user is None:
        response.status_code = 400
        return {'d': error_dict('api_errors', 'no valid username provided')}

    salted_pass = binascii.b2a_hex(hash_password(password, user.salt))

    if binascii.b2a_hex(user.password) != salted_pass:
        response.status_code = 400
        return {'d': error_dict('api_errors', 'no valid username provided')}

    # Account locking is not implemented in this version

    new_token = str(uuid4())

    new_session = SessionORM()
    new_session.user_id = user.user_id
    new_session.token = new_token
    request.state.dbsession.add(new_session)
    request.state.dbsession.flush()
    request.state.dbsession.refresh(new_session)

    since = datetime.now() - timedelta(weeks=2)
    request.state.dbsession.query(SessionORM).filter(
        SessionORM.updated_at <= since
    ).filter(SessionORM.user_id == user.user_id).delete()

    result = dict_from_row(new_session)
    result['origin'] = user.origin
    # Add verifier
    result['verifier'] = generate_verifier(str(user.user_id), user.salt, new_token)
    return {'d': result}


@sessionsRouter.delete('/sessions', summary='delete session')
async def sessions_delete_view(request: Request, response: Response):
    json_body = await request.json()
    if request.state.user is None:
        response.status_code = 400
        return {'d': error_dict('api_errors', 'not authenticated for this request')}
    token = json_body.get('token')
    if token is None or not isinstance(token, str):
        response.status_code = 400
        return {'d': error_dict('api_errors', 'no valid token provided')}
    s = (
        request.state.dbsession.query(SessionORM)
        .filter(SessionORM.token == token)
        .one_or_none()
    )
    if s is None:
        response.status_code = 400
        return {'d': error_dict('api_errors', 'no valid token provided')}

    request.state.dbsession.delete(s)
    request.state.dbsession.flush()

    return {'d': {}}


@sessionsRouter.put('/sessions', summary='validate session')
async def sessions_put_view(request: Request, response: Response):
    # This endpoint is for validating a session token when the client is not authenticated
    # The user middleware will attempt to authenticate with the token first
    if request.state.user is None:
        response.status_code = 400
        return {'d': error_dict('api_errors', 'not authenticated for this request')}

    json_body = await request.json()
    token = json_body.get('token')
    if token is None or not isinstance(token, str):
        response.status_code = 400
        return {'d': error_dict('api_errors', 'no valid token provided')}

    # Get the session matching both the token and the authenticated user
    s = (
        request.state.dbsession.query(SessionORM)
        .filter(
            SessionORM.token == token, SessionORM.user_id == request.state.user.user_id
        )
        .one_or_none()
    )
    if s is None:
        response.status_code = 400
        return {'d': error_dict('api_errors', 'no valid token provided')}

    expiration_value = timedelta(weeks=2)
    if (datetime.now() - s.updated_at) > expiration_value:
        request.state.dbsession.delete(s)
        response.status_code = 400
        return {'d': error_dict('api_errors', 'no valid token provided')}

    # Update the updated_at timestamp by explicitly setting it
    s.updated_at = datetime.now()
    request.state.dbsession.flush()

    result = dict_from_row(s)
    result['origin'] = request.state.user.origin
    # Add verifier
    result['verifier'] = generate_verifier(
        str(request.state.user.user_id), request.state.user.salt, s.token
    )
    return {'d': result}


@sessionsRouter.post('/sessions/rotate', summary='rotate session token')
async def sessions_rotate_view(request: Request, response: Response):
    """
    Rotate a session token. If a session is 13-14 days old, it marks the current session
    as rotated and creates a new one. Otherwise, returns an error.

    If rotate_early=True is passed, sessions younger than 14 days will be rotated, but
    the previous session will be deleted entirely rather than marked as rotated.

    The following cases are handled:
    - If the session token is not provided, return a 400 Bad Request
    - If the session token is invalid, return a 403 Forbidden
    - If the session is not 13-14 days old and rotate_early is not True, return a 401 Unauthorized
    - If the session is older than 14 days and rotate_early is True, return a 403 Forbidden
    - If the session is 13-14 days old, rotate it and return a 200 OK with the new session token
    - If the session is less than 14 days old and rotate_early is True, delete it and create a new one
    """
    json_body = await request.json()

    # Check if token is provided
    token = json_body.get('token')
    if token is None or not isinstance(token, str):
        response.status_code = status.HTTP_400_BAD_REQUEST
        return {'d': error_dict('api_errors', 'no token given')}

    # Get the rotate_early flag
    rotate_early = json_body.get('rotate_early', False)
    if not isinstance(rotate_early, bool):
        rotate_early = False

    # Look for the session with the given token
    session = (
        request.state.dbsession.query(SessionORM)
        .filter(SessionORM.token == token)
        .one_or_none()
    )

    # Check if session exists
    if session is None:
        response.status_code = status.HTTP_403_FORBIDDEN
        return {'d': error_dict('api_errors', 'bad token')}

    # Get the user from the session
    user = (
        request.state.dbsession.query(UserORM)
        .filter(UserORM.user_id == session.user_id)
        .one_or_none()
    )

    if user is None:
        response.status_code = status.HTTP_403_FORBIDDEN
        return {'d': error_dict('api_errors', 'bad token')}

    # Set rotation window (13-14 days)
    min_age = datetime.now() - timedelta(days=14)
    max_age = datetime.now() - timedelta(days=13)

    # Check if rotate_early is True and session is less than 14 days old
    session_age = datetime.now() - session.created_at

    if rotate_early:
        # If rotate_early is True but session is too old (> 14 days), return 403
        if session_age > timedelta(days=14):
            response.status_code = status.HTTP_403_FORBIDDEN
            return {'d': error_dict('api_errors', 'session not eligible for rotation')}
    else:
        # Standard rotation check - must be in 13-14 day window
        if not (min_age <= session.created_at <= max_age):
            response.status_code = status.HTTP_401_UNAUTHORIZED
            return {'d': error_dict('api_errors', 'session not eligible for rotation')}

    # Handle session rotation based on rotate_early flag
    if rotate_early and session_age <= timedelta(days=14):
        # Delete the old session entirely for early rotation
        request.state.dbsession.delete(session)
        request.state.dbsession.flush()
    else:
        # Mark the current session as rotated for standard rotation
        session.rotated = True
        request.state.dbsession.flush()

    # Create a new session
    new_token = str(uuid4())
    new_session = SessionORM()
    new_session.user_id = user.user_id
    new_session.token = new_token
    new_session.rotated = False

    request.state.dbsession.add(new_session)
    request.state.dbsession.flush()
    request.state.dbsession.refresh(new_session)

    # Return the new session
    result = dict_from_row(new_session)
    if hasattr(user, 'origin'):
        result['origin'] = user.origin

    # Add verifier
    result['verifier'] = generate_verifier(str(user.user_id), user.salt, new_token)

    return {'d': result}
