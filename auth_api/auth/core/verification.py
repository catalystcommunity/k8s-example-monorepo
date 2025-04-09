import hashlib
import binascii
import json
from typing import Optional
from fastapi import Request, Response, status
from starlette.middleware.base import BaseHTTPMiddleware

from auth.core.error_dict import error_dict
from auth.core.models import UserORM, SessionORM


def generate_verifier(user_id: str, salt: bytes, token: str) -> str:
    """
    Generate a verifier hash as a SHA256 of user_id string + salt hex string + token string.

    Args:
        user_id: The user ID as a string
        salt: The user's salt as bytes
        token: The session token as a string

    Returns:
        A hex string (uppercase) of the SHA256 hash
    """
    # Convert salt to uppercase hex string
    salt_hex = binascii.b2a_hex(salt).decode('ascii').upper()

    # Concatenate values in specified order
    combined = f'{user_id}{salt_hex}{token}'

    # Generate SHA256 hash and return uppercase hex string
    hash_obj = hashlib.sha256(combined.encode('utf-8'))
    return hash_obj.hexdigest().upper()


def verify_session(token: str, verifier: str, user_id: str, salt: bytes) -> bool:
    """
    Verify that a session token and verifier match for a given user.

    Args:
        token: The session token as a string
        verifier: The provided verifier string to check
        user_id: The user ID as a string
        salt: The user's salt as bytes

    Returns:
        True if the verifier is correct, False otherwise
    """
    expected_verifier = generate_verifier(user_id, salt, token)
    return verifier == expected_verifier


class VerificationMiddleware(BaseHTTPMiddleware):
    """
    Middleware that validates the verifier in requests with tokens.

    This middleware:
    1. Sets request.state.verified = False by default
    2. If no token is provided in the request, sets request.state.user = None and continues
    3. If a token is provided but no verifier, returns a 401 unauthorized error
    4. If a token and verifier are provided but verification fails, returns a 401 unauthorized error
    """

    async def dispatch(self, request: Request, call_next):
        """Process each request to verify tokens and verifiers."""
        # Default verified state
        request.state.verified = False

        # For GET requests, check query parameters
        request_token = None
        request_verifier = None

        if request.method == 'GET':
            request_token = request.query_params.get('token')
            request_verifier = request.query_params.get('verifier')
        # For PUT, POST, DELETE requests, check JSON body
        elif request.method in ['PUT', 'POST', 'DELETE']:
            # Skip if no body
            if len(await request.body()) == 0:
                request.state.user = None
                response = await call_next(request)
                return response

            try:
                json_body = await request.json()
                request_token = json_body.get('token')
                request_verifier = json_body.get('verifier')
            except json.decoder.JSONDecodeError:
                # Invalid JSON, let the endpoint handle it
                request.state.user = None
                response = await call_next(request)
                return response

        # If no token provided, user is None and continue
        if not request_token:
            request.state.user = None
            response = await call_next(request)
            return response

        # If token is provided but no verifier, return 401
        if request_token and not request_verifier:
            response = Response(
                content=json.dumps(
                    {'d': error_dict('api_errors', 'no verifier given')}
                ),
                media_type='application/json',
            )
            response.status_code = status.HTTP_401_UNAUTHORIZED
            return response

        # Look up user and session for verification
        session = (
            request.state.dbsession.query(SessionORM)
            .filter(SessionORM.token == request_token)
            .one_or_none()
        )

        # If session not found, continue with user=None
        if not session:
            request.state.user = None
            response = await call_next(request)
            return response

        # Get user from session
        user = (
            request.state.dbsession.query(UserORM)
            .filter(UserORM.id == session.user_id)
            .one_or_none()
        )

        # If user not found, continue with user=None
        if not user:
            request.state.user = None
            response = await call_next(request)
            return response

        # Verify the verifier
        is_verified = verify_session(
            request_token, request_verifier, str(user.id), user.salt
        )

        # If verification fails, return 401
        if not is_verified:
            response = Response(
                content=json.dumps(
                    {'d': error_dict('api_errors', 'session cannot be verified')}
                ),
                media_type='application/json',
            )
            response.status_code = status.HTTP_401_UNAUTHORIZED
            return response

        # Set verified state and user if verification passes
        request.state.verified = True
        request.state.user = user

        # Continue with the request
        response = await call_next(request)
        return response
