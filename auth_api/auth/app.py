# -*- coding: utf-8 -*-
from starlette.middleware import Middleware
from fastapi import FastAPI

from auth.core.db import TransactionMiddleware
from auth.core.logger import auth_logger
from auth.core.env import get_env_var
from auth.core.verification import VerificationMiddleware

from auth.views.base_views import baseRouter
from auth.views.session_views import sessionsRouter
from auth.views.user_views import usersRouter


APP_NAME = get_env_var('APP_NAME', 'auth')

App = FastAPI(
    title=APP_NAME,
    # We dont use app.middleware because these middlewares are not independent.
    # And the ordering here is easier to understand and consistent.
    middleware=[
        Middleware(TransactionMiddleware, logger=auth_logger),
        Middleware(VerificationMiddleware),
    ],
)

App.include_router(baseRouter)
App.include_router(sessionsRouter)
App.include_router(usersRouter)


# The following is all for intermediate testing purposes only, once we are ready to assign role based authentication
# to our actual endpoints all of this testing code can be refactored to use those views, though these views are
# entirely inaccessible to the web as they have no route, and they have no effects, so this doesn't cause any issues
# just leaving them in


# @authorized_roles()
# def unsecured_view(request):
#     return {}


# @authorized_roles(['user'])
# def user_view(request):
#     return {}


# @authorized_roles(['NonexistentRole'])
# def forbidden_view(request):
#     return {}


# @authorized_roles(['admin'])
# def admin_role_view(request):
#     return {}