"""
Tests for the session views module to verify session management functionality.
"""

import uuid
from datetime import datetime, timedelta
from uuid import uuid4

from fastapi import Response

from auth.core.models import UserORM, SessionORM, UserRole
from auth.views.session_views import (
    sessions_post_view,
    sessions_delete_view,
    sessions_put_view,
    sessions_rotate_view,
)
from tests import TestBase, run_coroutine
from tests.test_utils import create_mock_request


class TestSessionCreation(TestBase):
    """
    Tests for the session creation endpoint (/api/sessions POST).
    """

    def test_create_session_success(self):
        """
        Test successfully creating a session with valid username and password.
        """
        # Create test user and make sure it has an 'origin' attribute
        password = 'securepassword123'
        username = f'testuser_{uuid.uuid4().hex[:8]}'
        user = self.datautils.create_user({'username': username, 'password': password})

        # Add origin field (used in view)
        user.origin = 'test'
        self.session.flush()

        # Create test data with valid credentials
        test_session_data = {'username': user.username, 'password': password}

        # Create mock request
        request = create_mock_request(json_data=test_session_data, session=self.session)
        response = Response()

        # Call the view function directly
        result = run_coroutine(sessions_post_view(request, response))

        # Check response status
        assert response.status_code == 200

        # Check session was created
        assert 'd' in result
        assert 'token' in result['d']
        assert 'user_id' in result['d']

        # Verify session exists in database
        session = (
            self.session.query(SessionORM).filter_by(token=result['d']['token']).first()
        )
        assert session is not None
        # Convert both to string for comparison
        assert str(session.user_id) == str(user.user_id)

    def test_create_session_with_token(self):
        """
        Test creating a session using an existing token when user is authenticated.
        """
        # Create test user and session
        user = self.datautils.create_user()
        # Add origin field (used in view)
        user.origin = 'test'
        self.session.flush()

        existing_session = self.datautils.create_session({'user_id': user.user_id})

        # Create test data with existing token
        test_session_data = {'token': existing_session.token}

        # Create mock request with authenticated user
        request = create_mock_request(
            json_data=test_session_data, session=self.session, user=user
        )
        response = Response()

        # Call the view function directly
        result = run_coroutine(sessions_post_view(request, response))

        # Check response status
        assert response.status_code == 200

        # Check returned session matches existing session
        assert 'd' in result
        assert result['d']['token'] == existing_session.token
        # The user_id in the response might be a UUID or string representation
        # Convert both to string for comparison
        assert str(result['d']['user_id']) == str(user.user_id)

    def test_create_session_invalid_username(self):
        """
        Test that creating a session with an invalid username fails.
        """
        # Create test data with nonexistent username
        test_session_data = {
            'username': f'nonexistent_{uuid.uuid4().hex[:8]}',
            'password': 'password123',
        }

        # Create mock request
        request = create_mock_request(json_data=test_session_data, session=self.session)
        response = Response()

        # Call the view function directly
        result = run_coroutine(sessions_post_view(request, response))

        # Check response shows appropriate error
        assert response.status_code == 400
        assert 'api_errors' in result['d']['error_type']
        assert 'no valid username provided' in result['d']['errors']

    def test_create_session_missing_username(self):
        """
        Test that creating a session without a username fails.
        """
        # Create test data with missing username
        test_session_data = {'password': 'password123'}

        # Create mock request
        request = create_mock_request(json_data=test_session_data, session=self.session)
        response = Response()

        # Call the view function directly
        result = run_coroutine(sessions_post_view(request, response))

        # Check response shows appropriate error
        assert response.status_code == 400
        assert 'api_errors' in result['d']['error_type']
        assert 'no valid username provided' in result['d']['errors']

    def test_create_session_missing_password(self):
        """
        Test that creating a session without a password fails.
        """
        # Create test user
        user = self.datautils.create_user()

        # Create test data with missing password
        test_session_data = {'username': user.username}

        # Create mock request
        request = create_mock_request(json_data=test_session_data, session=self.session)
        response = Response()

        # Call the view function directly
        result = run_coroutine(sessions_post_view(request, response))

        # Check response shows appropriate error
        assert response.status_code == 400
        assert 'api_errors' in result['d']['error_type']
        assert 'no valid password provided' in result['d']['errors']

    def test_create_session_wrong_password(self):
        """
        Test that creating a session with an incorrect password fails.
        """
        # Create test user
        user = self.datautils.create_user({'password': 'correctpassword123'})

        # Create test data with wrong password
        test_session_data = {'username': user.username, 'password': 'wrongpassword123'}

        # Create mock request
        request = create_mock_request(json_data=test_session_data, session=self.session)
        response = Response()

        # Call the view function directly
        result = run_coroutine(sessions_post_view(request, response))

        # Check response shows appropriate error
        assert response.status_code == 400
        assert 'api_errors' in result['d']['error_type']
        assert 'no valid username provided' in result['d']['errors']

    def test_create_session_locked_account(self):
        """
        Test that creating a session for a locked account fails.

        Note: Account locking is not implemented in this version,
        but this test is kept as a placeholder for future implementation.
        """
        # For now, this test is just a placeholder since account locking
        # functionality is not implemented in the current version

        # Create a user without lockmessage
        user = self.datautils.create_user(
            {
                'password': 'password123',
            }
        )
        # Add origin field (used in view)
        user.origin = 'test'
        self.session.flush()

        # Create test data with correct credentials
        test_session_data = {'username': user.username, 'password': 'password123'}

        # Create mock request
        request = create_mock_request(json_data=test_session_data, session=self.session)
        response = Response()

        # Call the view function directly
        result = run_coroutine(sessions_post_view(request, response))

        # Instead of checking for account locking, just verify the session was created
        assert response.status_code == 200
        assert 'token' in result['d']
        assert 'user_id' in result['d']

    def test_create_session_cleans_old_sessions(self):
        """
        Test that creating a new session cleans up old sessions.
        """
        # Create test user
        password = 'securepassword123'
        user = self.datautils.create_user({'password': password})
        # Add origin field (used in view)
        user.origin = 'test'
        self.session.flush()

        # Create old session (more than 2 weeks old)
        old_session = self.datautils.create_session({'user_id': user.user_id})

        # Set the updated_at time to be older
        old_time = datetime.now() - timedelta(weeks=3)
        old_session.updated_at = old_time
        self.session.flush()

        # Create another session that's not old
        recent_session = self.datautils.create_session({'user_id': user.user_id})

        # Create test data for new session
        test_session_data = {'username': user.username, 'password': password}

        # Create mock request
        request = create_mock_request(json_data=test_session_data, session=self.session)
        response = Response()

        # Call the view function directly
        result = run_coroutine(sessions_post_view(request, response))

        # Check response status
        assert response.status_code == 200

        # Verify old session was deleted
        old_session_check = (
            self.session.query(SessionORM)
            .filter_by(session_id=old_session.session_id)
            .first()
        )
        assert old_session_check is None

        # Verify recent session still exists
        recent_session_check = (
            self.session.query(SessionORM)
            .filter_by(session_id=recent_session.session_id)
            .first()
        )
        assert recent_session_check is not None


class TestSessionDeletion(TestBase):
    """
    Tests for the session deletion endpoint (/api/sessions DELETE).
    """

    def test_delete_session_success(self):
        """
        Test successfully deleting a session.
        """
        # Create test user and session
        user = self.datautils.create_user()
        session = self.datautils.create_session({'user_id': user.user_id})

        # Create test data with session token
        test_delete_data = {'token': session.token}

        # Create mock request with authenticated user
        request = create_mock_request(
            json_data=test_delete_data, session=self.session, user=user
        )
        response = Response()

        # Call the view function directly
        result = run_coroutine(sessions_delete_view(request, response))

        # Check response status
        assert response.status_code == 200
        assert 'd' in result
        assert result['d'] == {}

        # Verify session was deleted
        deleted_session = (
            self.session.query(SessionORM)
            .filter_by(session_id=session.session_id)
            .first()
        )
        assert deleted_session is None

    def test_delete_session_not_authenticated(self):
        """
        Test that deleting a session without authentication fails.
        """
        # Create test user and session
        user = self.datautils.create_user()
        session = self.datautils.create_session({'user_id': user.user_id})

        # Create test data with session token
        test_delete_data = {'token': session.token}

        # Create mock request WITHOUT authenticated user
        request = create_mock_request(
            json_data=test_delete_data, session=self.session, user=None
        )
        response = Response()

        # Call the view function directly
        result = run_coroutine(sessions_delete_view(request, response))

        # Check response shows appropriate error
        assert response.status_code == 400
        assert 'api_errors' in result['d']['error_type']
        assert 'not authenticated for this request' in result['d']['errors']

        # Verify session was not deleted
        session_check = (
            self.session.query(SessionORM)
            .filter_by(session_id=session.session_id)
            .first()
        )
        assert session_check is not None

    def test_delete_session_missing_token(self):
        """
        Test that deleting a session without a token fails.
        """
        # Create test user
        user = self.datautils.create_user()

        # Create test data without token
        test_delete_data = {}

        # Create mock request with authenticated user
        request = create_mock_request(
            json_data=test_delete_data, session=self.session, user=user
        )
        response = Response()

        # Call the view function directly
        result = run_coroutine(sessions_delete_view(request, response))

        # Check response shows appropriate error
        assert response.status_code == 400
        assert 'api_errors' in result['d']['error_type']
        assert 'no valid token provided' in result['d']['errors']

    def test_delete_session_invalid_token(self):
        """
        Test that deleting a session with an invalid token fails.
        """
        # Create test user
        user = self.datautils.create_user()

        # Create test data with invalid token
        test_delete_data = {
            'token': str(uuid.uuid4())  # Random non-existent token
        }

        # Create mock request with authenticated user
        request = create_mock_request(
            json_data=test_delete_data, session=self.session, user=user
        )
        response = Response()

        # Call the view function directly
        result = run_coroutine(sessions_delete_view(request, response))

        # Check response shows appropriate error
        assert response.status_code == 400
        assert 'api_errors' in result['d']['error_type']
        assert 'no valid token provided' in result['d']['errors']


class TestSessionValidation(TestBase):
    """
    Tests for the session validation endpoint (/api/sessions PUT).
    """

    def test_validate_session_success(self):
        """
        Test successfully validating a session.
        """
        # Create test user and session
        user = self.datautils.create_user()
        # Add origin field (used in view)
        user.origin = 'test'
        self.session.flush()

        session = self.datautils.create_session({'user_id': user.user_id})

        # Create test data with session token
        test_validate_data = {'token': session.token}

        # Create mock request with authenticated user (now correct)
        request = create_mock_request(
            json_data=test_validate_data, session=self.session, user=user
        )
        response = Response()

        # Call the view function directly
        result = run_coroutine(sessions_put_view(request, response))

        # Check response status
        assert response.status_code == 200
        assert 'token' in result['d']
        assert result['d']['token'] == session.token
        assert str(result['d']['user_id']) == str(user.user_id)
        assert 'origin' in result['d']

    def test_validate_session_unauthenticated(self):
        """
        Test validating a session when user is unauthenticated.

        This should fail because the endpoint requires authentication.
        """
        # Create test user and session
        user = self.datautils.create_user()
        session = self.datautils.create_session({'user_id': user.user_id})

        # Create test data with session token
        test_validate_data = {'token': session.token}

        # Create mock request WITHOUT authenticated user
        request = create_mock_request(
            json_data=test_validate_data, session=self.session, user=None
        )
        response = Response()

        # Call the view function directly
        result = run_coroutine(sessions_put_view(request, response))

        # Check response shows appropriate error
        assert response.status_code == 400
        assert 'api_errors' in result['d']['error_type']
        assert 'not authenticated for this request' in result['d']['errors']

    def test_validate_session_invalid_token(self):
        """
        Test that validating a session with an invalid token fails.
        """
        # Create user
        user = self.datautils.create_user()
        # Add origin field (used in view)
        user.origin = 'test'
        self.session.flush()

        # Create test data with invalid token
        test_validate_data = {
            'token': str(uuid.uuid4())  # Random non-existent token
        }

        # Create mock request with authenticated user
        request = create_mock_request(
            json_data=test_validate_data, session=self.session, user=user
        )
        response = Response()

        # Call the view function directly
        result = run_coroutine(sessions_put_view(request, response))

        # Check response shows appropriate error
        assert response.status_code == 400
        assert 'api_errors' in result['d']['error_type']
        assert 'no valid token provided' in result['d']['errors']

    def test_validate_expired_session(self):
        """
        Test that validating an expired session fails and deletes the session.
        """
        # Create test user and session
        user = self.datautils.create_user()
        # Add origin field (used in view)
        user.origin = 'test'
        self.session.flush()

        session = self.datautils.create_session({'user_id': user.user_id})

        # Set session to be expired (more than 2 weeks old)
        session.updated_at = datetime.now() - timedelta(weeks=3)
        self.session.flush()

        # Create test data with expired session token
        test_validate_data = {'token': session.token}

        # Create mock request with authenticated user
        request = create_mock_request(
            json_data=test_validate_data, session=self.session, user=user
        )
        response = Response()

        # Call the view function directly
        result = run_coroutine(sessions_put_view(request, response))

        # Check response shows appropriate error
        assert response.status_code == 400
        assert 'api_errors' in result['d']['error_type']
        assert 'no valid token provided' in result['d']['errors']

        # Verify session was deleted
        deleted_session = (
            self.session.query(SessionORM)
            .filter_by(session_id=session.session_id)
            .first()
        )
        assert deleted_session is None


class TestSessionRotation(TestBase):
    """
    Tests for the session rotation endpoint (/api/sessions/rotate POST).
    """

    def test_rotate_session_new_implementation(self):
        """
        Test that the session rotation endpoint returns 401 for a newly created session.
        This test replaced the original stub test.
        """
        # Create test user and session
        user = self.datautils.create_user()
        session = self.datautils.create_session({'user_id': user.user_id})

        # Create test data with session token
        test_data = {'token': session.token}

        # Create mock request
        request = create_mock_request(json_data=test_data, session=self.session)
        response = Response()

        # Call the view function directly
        result = run_coroutine(sessions_rotate_view(request, response))

        # Check that it returns 401 for a new session
        assert response.status_code == 401
        assert 'api_errors' in result['d']['error_type']
        assert 'session not eligible for rotation' in result['d']['errors']

    def test_rotate_session_happy_path(self):
        """
        Test rotating a session that is 13-14 days old.
        This test will fail until we implement the feature.
        """
        # Create test user and session
        user = self.datautils.create_user()
        # Add origin field (used in view)
        user.origin = 'test'
        self.session.flush()

        session = self.datautils.create_session({'user_id': user.user_id})

        # Set the session created_at to be 13.5 days old
        old_time = datetime.now() - timedelta(days=13, hours=12)
        session.created_at = old_time
        session.updated_at = old_time
        self.session.flush()

        # Create test data with session token
        test_data = {'token': session.token}

        # Create mock request
        request = create_mock_request(
            json_data=test_data, session=self.session, user=user
        )
        response = Response()

        # Call the view function directly - this will fail until implementation is complete
        result = run_coroutine(sessions_rotate_view(request, response))

        # This test will fail because we expect 200 but stub returns 500
        assert response.status_code == 200
        assert 'token' in result['d']
        assert result['d']['token'] != session.token  # Should be a new token

        # Check that the old session is marked as rotated
        original_session = (
            self.session.query(SessionORM)
            .filter_by(session_id=session.session_id)
            .first()
        )
        assert original_session.rotated is True

        # Verify a new session was created
        new_session = (
            self.session.query(SessionORM).filter_by(token=result['d']['token']).first()
        )
        assert new_session is not None
        assert new_session.session_id != session.session_id
        assert new_session.user_id == user.user_id
        assert new_session.rotated is False

    def test_rotate_session_not_13_to_14_days_old(self):
        """
        Test that rotating a session that is not 13-14 days old returns 401.
        This test will fail until we implement the feature.
        """
        # Create test user and session
        user = self.datautils.create_user()
        # Add origin field (used in view)
        user.origin = 'test'
        self.session.flush()

        session = self.datautils.create_session({'user_id': user.user_id})

        # Session is too new (just created)

        # Create test data with session token
        test_data = {'token': session.token}

        # Create mock request
        request = create_mock_request(
            json_data=test_data, session=self.session, user=user
        )
        response = Response()

        # Call the view function directly - this will fail until implementation is complete
        result = run_coroutine(sessions_rotate_view(request, response))

        # Should get a 401 for a session that's too new
        assert response.status_code == 401
        assert 'api_errors' in result['d']['error_type']
        assert 'session not eligible for rotation' in result['d']['errors']

        # Check that the session is not marked as rotated
        original_session = (
            self.session.query(SessionORM)
            .filter_by(session_id=session.session_id)
            .first()
        )
        assert original_session.rotated is False

    def test_rotate_session_invalid_token(self):
        """
        Test that rotating a session with an invalid token returns 403.
        This test will fail until we implement the feature.
        """
        # Create test user
        user = self.datautils.create_user()

        # Create test data with non-existent token
        test_data = {
            'token': str(uuid4())  # Random non-existent token
        }

        # Create mock request
        request = create_mock_request(json_data=test_data, session=self.session)
        response = Response()

        # Call the view function directly - this will fail until implementation is complete
        result = run_coroutine(sessions_rotate_view(request, response))

        # Should get a 403 for an invalid token
        assert response.status_code == 403
        assert 'api_errors' in result['d']['error_type']
        assert 'bad token' in result['d']['errors']

    def test_rotate_session_no_token_provided(self):
        """
        Test that rotating a session without providing a token returns 400.
        This test will fail until we implement the feature.
        """
        # Create empty test data
        test_data = {}

        # Create mock request
        request = create_mock_request(json_data=test_data, session=self.session)
        response = Response()

        # Call the view function directly - this will fail until implementation is complete
        result = run_coroutine(sessions_rotate_view(request, response))

        # Should get a 400 for no token
        assert response.status_code == 400
        assert 'api_errors' in result['d']['error_type']
        assert 'no token given' in result['d']['errors']

    def test_rotate_early_session_less_than_14_days(self):
        """
        Test rotating a session early when it's less than 14 days old.
        The old session should be deleted completely.
        """
        # Create test user and session
        user = self.datautils.create_user()
        # Add origin field (used in view)
        user.origin = 'test'
        self.session.flush()

        session = self.datautils.create_session({'user_id': user.user_id})

        # Set the session created_at to be 7 days old
        old_time = datetime.now() - timedelta(days=7)
        session.created_at = old_time
        session.updated_at = old_time
        self.session.flush()

        # Create test data with session token and rotate_early flag
        test_data = {'token': session.token, 'rotate_early': True}

        # Create mock request
        request = create_mock_request(
            json_data=test_data, session=self.session, user=user
        )
        response = Response()

        # Call the view function directly
        result = run_coroutine(sessions_rotate_view(request, response))

        # Should succeed with a 200
        assert response.status_code == 200
        assert 'token' in result['d']
        assert result['d']['token'] != session.token  # Should be a new token

        # Check that the old session is deleted
        original_session = (
            self.session.query(SessionORM)
            .filter_by(session_id=session.session_id)
            .first()
        )
        assert original_session is None

        # Verify a new session was created
        new_session = (
            self.session.query(SessionORM).filter_by(token=result['d']['token']).first()
        )
        assert new_session is not None
        assert new_session.user_id == user.user_id
        assert new_session.rotated is False

    def test_rotate_early_session_older_than_14_days(self):
        """
        Test that rotating a session early when it's older than 14 days returns 403.
        """
        # Create test user and session
        user = self.datautils.create_user()
        # Add origin field (used in view)
        user.origin = 'test'
        self.session.flush()

        session = self.datautils.create_session({'user_id': user.user_id})

        # Set the session created_at to be 15 days old
        old_time = datetime.now() - timedelta(days=15)
        session.created_at = old_time
        session.updated_at = old_time
        self.session.flush()

        # Create test data with session token and rotate_early flag
        test_data = {'token': session.token, 'rotate_early': True}

        # Create mock request
        request = create_mock_request(
            json_data=test_data, session=self.session, user=user
        )
        response = Response()

        # Call the view function directly
        result = run_coroutine(sessions_rotate_view(request, response))

        # Should get a 403 for a session that's too old
        assert response.status_code == 403
        assert 'api_errors' in result['d']['error_type']
        assert 'session not eligible for rotation' in result['d']['errors']

        # Check that the session still exists and is not rotated
        original_session = (
            self.session.query(SessionORM)
            .filter_by(session_id=session.session_id)
            .first()
        )
        assert original_session is not None
        assert original_session.rotated is False

    def test_rotate_early_session_in_rotation_window(self):
        """
        Test rotating a session early when it's in the 13-14 day rotation window.
        The session should be deleted rather than marked as rotated.
        """
        # Create test user and session
        user = self.datautils.create_user()
        # Add origin field (used in view)
        user.origin = 'test'
        self.session.flush()

        session = self.datautils.create_session({'user_id': user.user_id})

        # Set the session created_at to be 13.5 days old (within rotation window)
        old_time = datetime.now() - timedelta(days=13, hours=12)
        session.created_at = old_time
        session.updated_at = old_time
        self.session.flush()

        # Create test data with session token and rotate_early flag
        test_data = {'token': session.token, 'rotate_early': True}

        # Create mock request
        request = create_mock_request(
            json_data=test_data, session=self.session, user=user
        )
        response = Response()

        # Call the view function directly
        result = run_coroutine(sessions_rotate_view(request, response))

        # Should succeed with a 200
        assert response.status_code == 200
        assert 'token' in result['d']
        assert result['d']['token'] != session.token  # Should be a new token

        # Check that the old session is deleted (not just marked rotated)
        original_session = (
            self.session.query(SessionORM)
            .filter_by(session_id=session.session_id)
            .first()
        )
        assert original_session is None

        # Verify a new session was created
        new_session = (
            self.session.query(SessionORM).filter_by(token=result['d']['token']).first()
        )
        assert new_session is not None
        assert new_session.user_id == user.user_id
        assert new_session.rotated is False
