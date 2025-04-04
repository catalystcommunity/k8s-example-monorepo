import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { getAllThingTypes, deleteThingType } from '../services/thingTypeService';

function ThingTypesList() {
  const [thingTypes, setThingTypes] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const { user } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    const fetchThingTypes = async () => {
      if (!user) return;
      
      try {
        setLoading(true);
        const data = await getAllThingTypes(user.token);
        setThingTypes(data);
        setError('');
      } catch (err) {
        setError('Failed to fetch thing types. ' + err.message);
      } finally {
        setLoading(false);
      }
    };

    fetchThingTypes();
  }, [user]);

  const handleEdit = (id) => {
    navigate(`/thing-types/edit/${id}`);
  };

  const handleDelete = async (id) => {
    if (window.confirm('Are you sure you want to delete this thing type? This may affect owned things that use this type.')) {
      try {
        await deleteThingType(id, user.token);
        setThingTypes(thingTypes.filter(type => type.thing_type_id !== id));
      } catch (err) {
        setError('Failed to delete. ' + err.message);
      }
    }
  };

  const handleCreate = () => {
    navigate('/thing-types/create');
  };

  if (loading) return <div>Loading...</div>;
  if (error) return <div className="error-message">{error}</div>;

  return (
    <div className="thing-types-container">
      <h1>Thing Types</h1>
      <button onClick={handleCreate} className="create-button">Create New Type</button>
      
      {thingTypes.length === 0 ? (
        <p>You haven't created any thing types yet. Click 'Create New Type' to add one.</p>
      ) : (
        <div className="thing-types-list">
          {thingTypes.map((type) => (
            <div key={type.thing_type_id} className="thing-type-card">
              <h3>{type.name}</h3>
              <p>Description: {type.description || 'No description provided'}</p>
              <p>Created: {new Date(type.created_at).toLocaleDateString()}</p>
              <div className="card-actions">
                <button onClick={() => handleEdit(type.thing_type_id)}>Edit</button>
                <button onClick={() => handleDelete(type.thing_type_id)} className="delete-button">Delete</button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

export default ThingTypesList;