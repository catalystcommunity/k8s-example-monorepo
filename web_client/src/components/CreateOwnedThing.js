import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { createOwnedThing } from '../services/ownedThingService';
import { getAllThingTypes } from '../services/thingTypeService';

function CreateOwnedThing() {
  const [name, setName] = useState('');
  const [thingTypeId, setThingTypeId] = useState('');
  const [thingTypes, setThingTypes] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const { user } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    const fetchThingTypes = async () => {
      if (!user) return;
      
      try {
        const data = await getAllThingTypes(user.token);
        setThingTypes(data);
        if (data.length > 0) {
          setThingTypeId(data[0].thing_type_id);
        } else {
          // Redirect to thing types if none exist
          setError('There are no thing types available. Please create some first.');
          setTimeout(() => {
            navigate('/thing-types');
          }, 2000);
        }
      } catch (err) {
        setError('Failed to fetch thing types. ' + err.message);
      }
    };

    fetchThingTypes();
  }, [user, navigate]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    
    if (!name.trim()) {
      setError('Name is required');
      return;
    }
    
    if (!thingTypeId) {
      setError('Thing type is required');
      return;
    }
    
    try {
      setLoading(true);
      await createOwnedThing({
        name,
        thing_type_id: thingTypeId,
        owner_id: user.user_id
      }, user.token);
      
      navigate('/owned-things');
    } catch (err) {
      setError('Failed to create owned thing. ' + err.message);
      setLoading(false);
    }
  };

  return (
    <div className="form-container">
      <h2>Create New Owned Thing</h2>
      {error && <p className="error-message">{error}</p>}
      
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="name">Name</label>
          <input
            type="text"
            id="name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />
        </div>
        
        <div className="form-group">
          <label htmlFor="thingType">Thing Type</label>
          <select
            id="thingType"
            value={thingTypeId}
            onChange={(e) => setThingTypeId(e.target.value)}
            required
          >
            {thingTypes.length === 0 && <option value="">No thing types available</option>}
            {thingTypes.map((type) => (
              <option key={type.thing_type_id} value={type.thing_type_id}>
                {type.name}
              </option>
            ))}
          </select>
        </div>
        
        <div className="form-actions">
          <button type="button" onClick={() => navigate('/owned-things')} className="cancel-button">
            Cancel
          </button>
          <button type="submit" disabled={loading}>
            {loading ? 'Creating...' : 'Create'}
          </button>
        </div>
      </form>
    </div>
  );
}

export default CreateOwnedThing;
