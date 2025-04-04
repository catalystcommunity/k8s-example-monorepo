import os

from auth.logger import auth_logger


# Retrieves a value from the environment, or returns a default value if it is not set, or None if no default is provided.
def get_env_var(name: str, default=None):
    return os.environ.get(name, default)


# Attempts to convert a value from the environment to an int and returns None if it fails, logging an error message.
def get_int_env_var(name: str, default=None):
    try:
        return int(os.environ.get(name, default))
    except ValueError:
        auth_logger.error(f"Failed to convert '{name}' to int")
        return None


def get_bool_env_var(name: str, default=None):
    return os.environ.get(name, default).lower() in ['true', '1', 'yes']
