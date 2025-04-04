import React, { createContext, useState, useEffect, useContext } from 'react';
import { login, signup, logout, validateSession } from '../services/authService';

const AuthContext = createContext();

export const useAuth = () => useContext(AuthContext);

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    // Check if there's a token in localStorage and validate it
    const checkAuth = async () => {
      const token = localStorage.getItem('token');
      if (token) {
        try {
          const userData = await validateSession(token);
          setUser(userData);
        } catch (err) {
          // Invalid or expired token
          localStorage.removeItem('token');
          localStorage.removeItem('user_id');
        }
      }
      setLoading(false);
    };

    checkAuth();
  }, []);

  const loginUser = async (username, password) => {
    setError('');
    try {
      const data = await login(username, password);
      setUser({
        user_id: data.user_id,
        token: data.token
      });
      localStorage.setItem('token', data.token);
      localStorage.setItem('user_id', data.user_id);
      return data;
    } catch (err) {
      setError(err.message || 'Failed to login');
      throw err;
    }
  };

  const signupUser = async (username, email, password) => {
    setError('');
    try {
      const data = await signup(username, email, password);
      setUser({
        user_id: data.user_id,
        token: data.session.token
      });
      localStorage.setItem('token', data.session.token);
      localStorage.setItem('user_id', data.user_id);
      return data;
    } catch (err) {
      setError(err.message || 'Failed to signup');
      throw err;
    }
  };

  const logoutUser = async () => {
    const token = localStorage.getItem('token');
    if (token) {
      try {
        await logout(token);
      } catch (err) {
        console.error('Error during logout:', err);
      }
      localStorage.removeItem('token');
      localStorage.removeItem('user_id');
      setUser(null);
    }
  };

  const value = {
    user,
    loading,
    error,
    loginUser,
    signupUser,
    logoutUser,
    isAuthenticated: !!user
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};
