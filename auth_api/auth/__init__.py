# -*- coding: utf-8 -*-
import datetime
import os
import json

from sqlalchemy import func, cast
from sqlalchemy.dialects.postgresql import INTERVAL
from fastapi import APIRouter, FastAPI, Request, Response, status, HTTPException

from starlette.middleware import Middleware
from starlette.middleware.base import BaseHTTPMiddleware

from auth.config import Config
from auth.db import TransactionMiddleware
from auth.models import User, Session
from auth.logging import auth_logger
from auth.env import get_env_var

from auth.views.base_views import baseRouter
from auth.views.session_views import sessionsRouter
from auth.views.user_views import usersRouter


APP_NAME = get_env_var('APP_NAME', 'auth')


class UserMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next):
        """
        This property will be added to the request, and for now it simply verifies if they
        are authenticated or not based on the token they provide.  If this fails, this property
        is None
        """
        request_token = 'invalid_token'
        # This defaults the state, showing intent, and also makes checking it easier to port.
        request.state.user = None
        if request.method == 'GET':
            if request.query_params.get('token') is None:
                response = await call_next(request)
                return response
            else:
                request_token = request.query_params.get('token')
        elif request.method in ['PUT', 'POST', 'DELETE']:
            # I'm not sure I like this, but it works.
            # I might just require a token to be provided in each endpoint
            if len(await request.body()) == 0:
                response = await call_next(request)
                return response
            try:
                json_body = await request.json()
            except json.decoder.JSONDecodeError:
                response = Response('Invalid JSON')
                response.status_code = status.HTTP_400_BAD_REQUEST
                return response

            if not hasattr(json_body, 'token'):
                response = await call_next(request)
                return response
            else:
                request_token = json_body.get('token')

        dauser = (
            request.state.dbsession.query(User)
            .filter(Session.token == request_token)
            .filter(
                Session.lastactive
                >= (func.current_timestamp() - cast('1 week', INTERVAL))
            )
            .join(Session, Session.user_id == User.id)
            .one_or_none()
        )
        if dauser:
            session = (
                request.state.dbsession.query(Session)
                .filter(Session.token == request_token)
                .one()
            )
            session.lastactive = datetime.datetime.now()
            request.state.dbsession.flush()

        request.state.user = dauser
        response = await call_next(request)
        return response


app = FastAPI(
    title=APP_NAME,
    # We dont use app.middleware because these two are not independent.
    # And the ordering here is easier to understand and consistent.
    middleware=[
        Middleware(TransactionMiddleware, logger=auth_logger),
        Middleware(UserMiddleware),
    ],
)

app.include_router(baseRouter)
app.include_router(sessionsRouter)
app.include_router(usersRouter)


# The following is all for intermediate testing purposes only, once we are ready to assign role based authentication
# to our actual endpoints all of this testing code can be refactored to use those views, though these views are
# entirely inaccessible to the web as they have no route, and they have no effects, so this doesn't cause any issues
# just leaving them in


# @authorized_roles()
# def unsecured_view(request):
#     return {}


# @authorized_roles(['MS_user'])
# def user_view(request):
#     return {}


# @authorized_roles(['NonexistentRole'])
# def forbidden_view(request):
#     return {}


# @authorized_roles(['MS_admin'])
# def admin_role_view(request):
#     return {}
