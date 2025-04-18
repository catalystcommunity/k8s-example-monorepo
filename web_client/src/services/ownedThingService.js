const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:3001';

// Helper function to handle API responses
const handleResponse = async (response) => {
  const data = await response.json();
  
  if (!response.ok) {
    const error = (data && data.error) || response.statusText;
    return Promise.reject(new Error(error));
  }
  
  return data.d || data;
};

// Get all owned things for a user
export const getOwnedThingsByOwner = async (ownerId, token) => {
  const response = await fetch(`${API_URL}/api/owned-things/owner/${ownerId}`, {
    method: 'GET',
    headers: { 
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    }
  });
  
  return handleResponse(response);
};

// Get a specific owned thing by ID
export const getOwnedThingById = async (id, token) => {
  const response = await fetch(`${API_URL}/api/owned-things/${id}`, {
    method: 'GET',
    headers: { 
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    }
  });
  
  return handleResponse(response);
};

// Create a new owned thing
export const createOwnedThing = async (ownedThingData, token) => {
  const response = await fetch(`${API_URL}/api/owned-things`, {
    method: 'POST',
    headers: { 
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify(ownedThingData)
  });
  
  return handleResponse(response);
};

// Update an existing owned thing
export const updateOwnedThing = async (id, updateData, token) => {
  const response = await fetch(`${API_URL}/api/owned-things/${id}`, {
    method: 'PUT',
    headers: { 
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify(updateData)
  });
  
  return handleResponse(response);
};

// Delete an owned thing
export const deleteOwnedThing = async (id, token) => {
  const response = await fetch(`${API_URL}/api/owned-things/${id}`, {
    method: 'DELETE',
    headers: { 
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    }
  });
  
  if (response.status === 204) {
    return { success: true };
  }
  
  return handleResponse(response);
};
