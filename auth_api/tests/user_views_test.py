"""
Tests for the user views module to verify user creation functionality.
"""

import uuid

from fastapi import Response

from auth.core.models import UserORM, UserRole
from auth.views.user_views import users_post_view
from tests.test_utils import create_mock_request, TestBase, run_coroutine

# Alias for backward compatibility in tests
User = UserORM


class TestUserCreation(TestBase):
    """
    Tests for the user creation endpoint (/api/users POST).
    """

    def test_create_user_success(self):
        """
        Test successfully creating a user with valid data.
        """
        # Create test data with unique values
        username = f'testuser_{uuid.uuid4().hex[:8]}'
        test_user_data = {
            'username': username,
            'email': f'{username}@example.com',
            'password': 'securepassword123',
        }

        # Create mock request with our test session
        request = create_mock_request(json_data=test_user_data, session=self.session)
        response = Response()

        # Call the view function directly
        result = run_coroutine(users_post_view(request, response))

        # Check response status
        assert response.status_code == 200

        # Check the user was created in the database
        user = self.session.query(UserORM).filter_by(username=username.lower()).first()
        assert user is not None
        assert user.username == username.lower()
        assert user.email == test_user_data['email'].lower()

        # Check password was properly hashed
        assert isinstance(user.password, bytes)
        assert isinstance(user.salt, bytes)

        # Verify default role was assigned
        assert user.roles == [UserRole.user]

        # Verify result contains user info and session
        assert 'd' in result
        assert result['d']['username'] == username.lower()
        assert result['d']['email'] == test_user_data['email'].lower()
        assert 'session' in result['d']
        assert 'token' in result['d']['session']

    def test_create_user_duplicate_username(self):
        """
        Test that creating a user with a duplicate username fails.
        """
        # First create a user with a unique username
        username = f'duplicate_{uuid.uuid4().hex[:8]}'
        self.datautils.create_user(
            {
                'username': username,
                'email': f'original_{username}@example.com',
            }
        )
        # Try to create another user with the same username but different email
        test_user_data = {
            'username': username,  # Same username as existing user
            'email': f'different_{username}@example.com',
            'password': 'differentpass123',
        }

        # Create mock request
        request = create_mock_request(json_data=test_user_data, session=self.session)
        response = Response()

        # Call the view function directly
        result = run_coroutine(users_post_view(request, response))

        # Check response shows appropriate error
        assert response.status_code == 400
        assert 'verification_error' in result['d']['error_type']
        assert any('username already in use' in err for err in result['d']['errors'])

    def test_create_user_missing_fields(self):
        """
        Test that creating a user with missing required fields fails.
        """
        test_cases = [
            # Missing password
            {'username': 'testuser1', 'email': 'testuser1@example.com'},
            # Missing email
            {'username': 'testuser2', 'password': 'password123'},
            # Missing username
            {'email': 'testuser3@example.com', 'password': 'password123'},
            # Empty request
            {},
        ]

        for test_data in test_cases:
            # Create mock request
            request = create_mock_request(json_data=test_data, session=self.session)
            response = Response()

            # Call the view function directly
            result = run_coroutine(users_post_view(request, response))

            # Check response shows error
            assert response.status_code == 400
            assert 'error_type' in result['d']
            assert 'api_errors' in result['d']['error_type']

    def test_create_user_invalid_types(self):
        """
        Test that creating a user with invalid field types fails.
        """
        # Username not a string
        test_user_data = {
            'username': 12345,  # Number instead of string
            'email': 'test@example.com',
            'password': 'password123',
        }

        # Create mock request
        request = create_mock_request(json_data=test_user_data, session=self.session)
        response = Response()

        # Call the view function directly
        result = run_coroutine(users_post_view(request, response))

        # Check response shows error
        assert response.status_code == 400
        assert 'error_type' in result['d']
        assert 'api_errors' in result['d']['error_type']

    def test_user_lowercase_conversion(self):
        """
        Test that usernames and emails are converted to lowercase.
        """
        # Create test data with mixed case
        test_user_data = {
            'username': 'TestUserMixedCase',
            'email': 'TestUser@Example.COM',
            'password': 'securepassword123',
        }

        # Create mock request with our test session
        request = create_mock_request(json_data=test_user_data, session=self.session)
        response = Response()

        # Call the view function directly
        result = run_coroutine(users_post_view(request, response))

        # Check response status
        assert response.status_code == 200

        # Check the user was created with lowercase values
        user = (
            self.session.query(UserORM).filter_by(username='testusermixedcase').first()
        )
        assert user is not None
        assert user.username == 'testusermixedcase'
        assert user.email == 'testuser@example.com'
