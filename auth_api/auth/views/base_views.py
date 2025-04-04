from fastapi import APIRouter, Request, HTTPException
from sqlalchemy import text
import time

baseRouter = APIRouter(prefix='/api', tags=['base'])

# def notfound_view(request):
#     """
#     For when Pyramid is passed a url it can not route
#     """
#     return HTTPNotFound('The url could not be found.')


@baseRouter.get('/health')
def health_check(request: Request):
    """
    Health check that verifies database connectivity by executing a lightweight query
    that tests the connection without reading from tables.

    This endpoint also returns the verification status if a token is provided.
    """
    try:
        # Get the database session from request state (added by TransactionMiddleware)
        db = request.state.dbsession

        # Execute the query to check database connectivity and uptime
        result = db.execute(
            text('SELECT current_timestamp - pg_postmaster_start_time() as uptime')
        ).scalar()

        # Convert to seconds for easier reading
        uptime_seconds = result.total_seconds()

        # Build response
        response = {
            'status': 'OK',
            'database': {'connected': True, 'uptime_seconds': uptime_seconds},
            'timestamp': time.time(),
        }

        # Add verification info if available
        if hasattr(request.state, 'verified'):
            response['verification'] = {
                'verified': request.state.verified,
                'user_authenticated': request.state.user is not None,
            }

        return response
    except Exception as e:
        # Log the error and return a 503 Service Unavailable
        raise HTTPException(
            status_code=503, detail=f'Database health check failed: {str(e)}'
        )


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
