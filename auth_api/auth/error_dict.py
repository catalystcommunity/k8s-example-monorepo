"""
The error dict for standard error responses.
"""


def error_dict(error_type='generic_errors', errors=''):
    """
    Create a basic error dictionary for standard use with the intent of being passed to some other outside
    API or whatnot.
    :param type: A plural string without spaces that describes the errors.  Only one type of error should be sent.
    :param errors: A list of error strings describing the problems. A single string will be converted to a single item
    list containing that string.
    :return: A dictionary for the error to be passed.
    """
    if isinstance(errors, str):
        errors = [errors]
    if not isinstance(errors, list):
        raise TypeError('Type for "errors" must be a string or list of strings')
    if not all(isinstance(item, str) for item in errors):
        raise TypeError('Type for "errors" in the list must be a string')
    error = {
        'error_type': error_type,
        'errors': errors,
    }
    return error
