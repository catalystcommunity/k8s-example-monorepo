from fastapi import APIRouter

baseRouter = APIRouter(prefix='/api', tags=['base'])

# def notfound_view(request):
#     """
#     For when Pyramid is passed a url it can not route
#     """
#     return HTTPNotFound('The url could not be found.')


@baseRouter.get('/health')
def health_view():
    """
    For a generalized health check
    """
    return {'status': 'OK'}


conn_err_msg = """\
Pyramid is having a problem using your SQL database.  The problem
might be caused by one of the following things:

1.  You may need to run the "initialize_utilities_db" script
    to initialize your database tables.  Check your virtual
    environment's "bin" directory for this script and try to run it.

2.  Your database server may not be running.  Check that the
    database server referred to by the "sqlalchemy.url" setting in
    your "development.ini" file is running.

After you fix the problem, please restart the Pyramid application to
try it again.
"""
