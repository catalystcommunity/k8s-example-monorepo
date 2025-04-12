from fastapi.testclient import TestClient

from auth.app import App

client = TestClient(App)


def test_health_endpoint():
    """
    Test the health check endpoint returns a successful response.
    """
    response = client.get('/api/health')
    assert response.status_code == 200
    assert 'status' in response.json()
    assert response.json()['status'] == 'OK'
    assert 'database' in response.json()
    assert response.json()['database']['connected'] is True
