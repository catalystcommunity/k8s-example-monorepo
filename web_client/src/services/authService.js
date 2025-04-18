const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:5080';

// Helper function to handle API responses
const handleResponse = async (response) => {
  const data = await response.json();
  
  if (!response.ok) {
    const error = (data && data.error) || response.statusText;
    return Promise.reject(new Error(error));
  }
  
  return data.d;
};

// Login user and get token
export const login = async (username, password) => {
  const response = await fetch(`${API_URL}/api/sessions`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password })
  });
  
  return handleResponse(response);
};

// Signup new user
export const signup = async (username, email, password) => {
  const response = await fetch(`${API_URL}/api/users`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, email, password })
  });
  
  return handleResponse(response);
};

// Logout user
export const logout = async (token) => {
  const response = await fetch(`${API_URL}/api/sessions`, {
    method: 'DELETE',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token })
  });
  
  return handleResponse(response);
};

// Validate session token
export const validateSession = async (token) => {
  const response = await fetch(`${API_URL}/api/sessions`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token })
  });
  
  return handleResponse(response);
};

// Rotate session token
export const rotateToken = async (token, rotateEarly = false) => {
  const response = await fetch(`${API_URL}/api/sessions/rotate`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token, rotate_early: rotateEarly })
  });
  
  return handleResponse(response);
};
