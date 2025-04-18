import React from 'react';
import { useAuth } from '../contexts/AuthContext';

function Home() {
  const { user, isAuthenticated } = useAuth();

  return (
    <div>
      <h1>Welcome to K8s Example App</h1>
      {isAuthenticated ? (
        <div>
          <p>You are logged in with user ID: {user.user_id}</p>
          <p>This is a simple demo application showing authentication flow.</p>
        </div>
      ) : (
        <div>
          <p>Please log in or sign up to access the application.</p>
        </div>
      )}
    </div>
  );
}

export default Home;
