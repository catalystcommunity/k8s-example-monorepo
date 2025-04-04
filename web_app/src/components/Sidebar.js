import React from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

function Sidebar() {
  const { isAuthenticated, logoutUser } = useAuth();

  return (
    <div className="sidebar">
      <h2>K8s Example App</h2>
      <ul className="nav-links">
        <li>
          <Link to="/">Home</Link>
        </li>
        {isAuthenticated ? (
          <>
            <li>
              <Link to="/owned-things">My Things</Link>
            </li>
            <li>
              <Link to="/thing-types">Thing Types</Link>
            </li>
            <li>
              <button onClick={logoutUser}>Logout</button>
            </li>
          </>
        ) : (
          <>
            <li>
              <Link to="/login">Login</Link>
            </li>
            <li>
              <Link to="/signup">Signup</Link>
            </li>
          </>
        )}
      </ul>
    </div>
  );
}

export default Sidebar;
