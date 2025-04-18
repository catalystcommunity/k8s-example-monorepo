# Web Frontend

A simple React frontend for the K8s Example Monorepo project.

## Features

- React-based SPA (Single Page Application)
- Authentication flow (login, signup, logout)
- Integration with auth_api
- Some basic shells for other views to get started with

## Development

```bash
# Install dependencies
npm install

# Start development server
npm start
```

## Environment Variables

- `REACT_APP_API_URL`: URL for API backend (default: 'http://localhost:4080')

If you run this through the docker image, nginx is the one serving it and you will need a fronting "proxy" to the app/auth APIs. That isn't included.

## Docker Build

```bash
docker build -t k8s-example-web-client .
```

## Project Structure

- `src/components/`: React components
- `src/contexts/`: Context providers (AuthContext)
- `src/services/`: API service functions
