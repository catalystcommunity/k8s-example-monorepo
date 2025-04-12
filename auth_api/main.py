import sys
import os
import platform
import uvicorn

from auth.core.config import Config
from auth.app import App

# Log environment info for debugging
print('Python paths:', sys.path)
print('Python version:', sys.version)
print('Python platform:', platform.platform())
print('Python implementation:', platform.python_implementation())
print('Environment variables:', os.environ.get('PYTHONPATH'))


def main():
    """
    Start the FastAPI application using Uvicorn
    """
    # Print startup information
    print(f'Starting auth service on 0.0.0.0:{Config.port}')
    print(f'Database URL: {Config.db_url}')

    # Start the server
    # Use host="0.0.0.0" to make the server accessible from outside the container
    uvicorn.run(App, host='0.0.0.0', port=Config.port)


if __name__ == '__main__':
    main()
