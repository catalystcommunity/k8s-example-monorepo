"""
Test utilities for auth_api tests.

This module provides helper functions and classes for tests.
"""


class MockRequest:
    """
    Mock request object that mimics FastAPI request with state and JSON data.

    This allows us to test view functions directly without spinning up
    a full web server.

    Args:
        json_data: The JSON data to return from the json() method
        dbsession: Database session to attach to request.state
        user: Optional user to attach to request.state (for authenticated endpoints)
    """

    def __init__(self, json_data, dbsession=None, user=None):
        self.json_data = json_data
        self.state = type('obj', (object,), {'dbsession': dbsession, 'user': user})
        self.query_params = {}
        self.path_params = {}
        self.headers = {}

    async def json(self):
        """Mock the FastAPI request.json() method"""
        return self.json_data

    async def body(self):
        """Mock the request.body() method"""
        return b''  # Empty body by default


def create_mock_request(
    json_data=None,
    session=None,
    user=None,
    query_params=None,
    path_params=None,
    headers=None,
):
    """
    Create a mock request with the given parameters.

    Args:
        json_data: JSON data for the request body
        session: Database session to use
        user: User for authentication
        query_params: Query parameters
        path_params: Path parameters
        headers: Request headers

    Returns:
        MockRequest object
    """
    if json_data is None:
        json_data = {}

    request = MockRequest(json_data, session, user)

    if query_params:
        request.query_params = query_params

    if path_params:
        request.path_params = path_params

    if headers:
        request.headers = headers

    return request
