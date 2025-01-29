import binascii
from datetime import datetime, timedelta
from uuid import uuid4
from fastapi import APIRouter, Request, Response

from auth.models import Session, User
from auth.error_dict import error_dict
from auth.passwords import hash_password

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
        return {
            'd': dict_from_row(
                request.state.dbsession.query(Session)
                .filter(Session.token == json_body['token'])
                .one()
            )
        }
    username = json_body.get('username')
    if username is None or not isinstance(username, str):
        response.status_code = 400
        return {'d': error_dict('api_errors', 'no valid username provided')}
    password = json_body.get('password')
    if password is None or not isinstance(password, str):
        response.status_code = 400
        return {'d': error_dict('api_errors', 'no valid password provided')}
    user = (
        request.state.dbsession.query(User)
        .filter(User.username == username.lower())
        .one_or_none()
    )
    if user is None:
        response.status_code = 400
        return {'d': error_dict('api_errors', 'no valid username provided')}

    salted_pass = binascii.b2a_hex(hash_password(password, user.salt))

    if binascii.b2a_hex(user.password) != salted_pass:
        response.status_code = 400
        return {'d': error_dict('api_errors', 'no valid username provided')}

    if user.lockmessage is not None:
        response.status_code = 400
        return {'d': error_dict('account_lock', user.lockmessage)}

    new_token = str(uuid4())

    new_session = Session()
    new_session.user_id = user.id
    new_session.token = new_token
    request.state.dbsession.add(new_session)
    request.state.dbsession.flush()
    request.state.dbsession.refresh(new_session)

    since = datetime.now() - timedelta(weeks=2)
    request.state.dbsession.query(Session).filter(Session.lastactive <= since).filter(
        Session.user_id == user.id
    ).delete()

    result = dict_from_row(new_session)
    result['origin'] = user.origin
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
        request.state.dbsession.query(Session)
        .filter(Session.token == token)
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
    if request.state.user is not None:
        response.status_code = 400
        return {'d': error_dict('api_errors', 'not authenticated for this request')}

    json_body = await request.json()
    token = json_body.get('token')
    s = (
        request.state.dbsession.query(Session)
        .filter(Session.token == token, Session.user_id == request.state.user.id)
        .one_or_none()
    )
    if s is None:
        response.status_code = 400
        return {'d': error_dict('api_errors', 'no valid token provided')}

    expiration_value = timedelta(weeks=2)
    if (datetime.now() - s.lastactive) > expiration_value:
        request.state.dbsession.delete(s)
        response.status_code = 400
        return {'d': error_dict('api_errors', 'no valid token provided')}

    s.lastactive = datetime.now()
    request.state.dbsession.flush()

    result = dict_from_row(s)
    result['origin'] = request.state.user.origin
    return {'d': result}
