# Web Frontend

A simple React frontend for the K8s Example Monorepo project.

## Features

- React-based SPA (Single Page Application)
- Authentication flow (login, signup, logout)
- Integration with auth_api
- Responsive UI

## Development

```bash
# Install dependencies
npm install

# Start development server
npm start
```

## Environment Variables

- `REACT_APP_API_URL`: URL for API backend (default: 'http://localhost:3001')

## Docker Build

```bash
docker build -t k8s-example-web-app .
```

## Project Structure

- `src/components/`: React components
- `src/contexts/`: Context providers (AuthContext)
- `src/services/`: API service functions
