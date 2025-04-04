import sys
import os
import platform

# Log environment info for debugging
print('Python paths:', sys.path)
print('Python version:', sys.version)
print('Python platform:', platform.platform())
print('Python implementation:', platform.python_implementation())
print('Environment variables:', os.environ.get('PYTHONPATH'))

# Import statements
import uvicorn
from auth.config import Config
from auth import app


def main():
    """
    Start the FastAPI application using Uvicorn
    """
    # Print startup information
    print(f'Starting auth service on 0.0.0.0:{Config.port}')
    print(f'Database URL: {Config.db_url}')

    # Start the server
    # Use host="0.0.0.0" to make the server accessible from outside the container
    uvicorn.run(app, host='0.0.0.0', port=Config.port)


if __name__ == '__main__':
    main()
