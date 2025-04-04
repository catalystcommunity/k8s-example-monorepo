# Tests for Auth API

This directory contains tests for the authentication API service.

## Overview

The tests for Auth API use pytest for running tests and use a real database for testing. Tests are wrapped in transactions that are rolled back after each test completes to ensure tests don't affect each other.

## Test Structure

- `setup_test.py` - Database setup and transaction fixtures
- `datautils.py` - Helper utilities for creating test data
- `test_utils.py` - Common test utilities like mock request objects
- `*_test.py` - Test modules for specific parts of the application

## Running Tests

To run the tests, make sure you have a database available and the correct connection string set in your environment or config files.

```bash
# Run all tests
pytest

# Run specific test file
pytest tests/user_views_test.py

# Run with coverage
pytest --cov=auth
```

## Test Approach

Tests are designed to be isolated and use real database transactions that are rolled back after each test. This approach provides several benefits:

1. Tests run against real database constraints and behavior
2. Tests don't interfere with each other
3. No need to mock database access
4. Tests are fast due to transaction rollback (no disk writes)

## Creating Test Data

The `DataUtils` class provides helper methods for creating test data with sensible defaults. Use this to create objects needed for your tests:

```python
def test_something(db_session, data_utils):
    # Create a user with default values
    user = data_utils.create_user()
    
    # Create a user with specific values
    custom_user = data_utils.create_user({
        "username": "custom_username",
        "email": "custom@example.com"
    })
    
    # Create a session for a user
    session = data_utils.create_session({"user_id": user.id})
```

## Testing API Endpoints

API endpoints can be tested by calling view functions directly with mock request objects:

```python
from tests.test_utils import create_mock_request
from auth.views.user_views import users_post_view

async def test_api_endpoint(db_session):
    request = create_mock_request(
        json_data={"field": "value"}, 
        session=db_session
    )
    response = Response()
    
    result = await users_post_view(request, response)
    
    assert response.status_code == 200
    assert "expected_field" in result
```

This approach allows testing endpoints without running a full web server while still exercising the full code path including database access.