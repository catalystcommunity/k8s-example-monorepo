const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:5080';

// Helper function to handle API responses
const handleResponse = async (response) => {
  const data = await response.json();
  
  if (!response.ok) {
    const error = (data && data.error) || response.statusText;
    return Promise.reject(new Error(error));
  }
  
  return data.d || data;
};

// Get all thing types
export const getAllThingTypes = async (token) => {
  const response = await fetch(`${API_URL}/api/thing-types`, {
    method: 'GET',
    headers: { 
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    }
  });
  
  return handleResponse(response);
};

// Get a specific thing type by ID
export const getThingTypeById = async (id, token) => {
  const response = await fetch(`${API_URL}/api/thing-types/${id}`, {
    method: 'GET',
    headers: { 
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    }
  });
  
  return handleResponse(response);
};

// Create a new thing type
export const createThingType = async (thingTypeData, token) => {
  const response = await fetch(`${API_URL}/api/thing-types`, {
    method: 'POST',
    headers: { 
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify(thingTypeData)
  });
  
  return handleResponse(response);
};

// Update an existing thing type
export const updateThingType = async (id, updateData, token) => {
  const response = await fetch(`${API_URL}/api/thing-types/${id}`, {
    method: 'PUT',
    headers: { 
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify(updateData)
  });
  
  return handleResponse(response);
};

// Delete a thing type
export const deleteThingType = async (id, token) => {
  const response = await fetch(`${API_URL}/api/thing-types/${id}`, {
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
