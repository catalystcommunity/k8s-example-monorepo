"""
Working with passwords should have a standard flow, here is those functions.
"""

import hashlib


def hash_password(password: str, salt: str) -> bytes:
    """
    Hashes a password with the salt and returns the hash
    :param password: a password string
    :type password: str
    :param salt: a salt hash
    :type salt: assumed to be bytes
    :return: a password digest, NOT a hex digest string
    """
    m = hashlib.sha512()
    m.update(password.encode('utf-8'))
    m.update(salt)
    return m.digest()
