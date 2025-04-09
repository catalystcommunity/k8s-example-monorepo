# A simple logger for the auth API module that logs messages to stdout
import json
import logging
import os
import sys
from pythonjsonlogger import json


_default_level = 'WARNING'
_new_level = os.environ.get('LOG_LEVEL', _default_level)


# Setup the logger for the auth module, this is the logger name that should be used in this whole package
auth_logger = logging.getLogger('auth_api')
log_format = '%(asctime)s - %(module)s#%(lineno)d - %(levelname)s - %(message)s'
formatter = json.JsonFormatter(
    fmt=log_format, rename_fields={'levelname': 'level', 'asctime': 'date'}
)

json_h = logging.StreamHandler(sys.stdout)
json_h.setFormatter(formatter)
auth_logger.addHandler(json_h)
auth_logger.setLevel(_default_level)

if _new_level in logging._nameToLevel:
    auth_logger.setLevel(_new_level)
else:
    logging.error(f'Invalid log level: {_new_level}, defaulting to {_default_level}')
