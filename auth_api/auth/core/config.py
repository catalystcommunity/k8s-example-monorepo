import os
from dataclasses import dataclass


# Don't modify this, but we don't limit that if you want to try to be "clever"
@dataclass()
class _config:
    pass


Config = _config()
# Set properties for config either from environ, or from some file/function parsing stuff
Config.db_url = os.environ.get(
    'DB_URI', 'postgresql://devuser:devpass@localhost/monodemopg'
)
Config.port = int(os.environ.get('PORT', '5080'))
Config.base_domain = os.environ.get('BASE_DOMAIN', f'localhost:{Config.port}')
