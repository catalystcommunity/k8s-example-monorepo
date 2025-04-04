import React, { useState, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { getOwnedThingsByOwner, deleteOwnedThing } from '../services/ownedThingService';
import { getAllThingTypes } from '../services/thingTypeService';

function OwnedThingsList() {
  const [ownedThings, setOwnedThings] = useState([]);
  const [thingTypes, setThingTypes] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const { user } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    const fetchData = async () => {
      if (!user) return;
      
      try {
        setLoading(true);
        
        // Use Promise.all to fetch both datasets in parallel
        const [ownedThingsData, thingTypesData] = await Promise.all([
          getOwnedThingsByOwner(user.user_id, user.token),
          getAllThingTypes(user.token)
        ]);
        
        setOwnedThings(ownedThingsData);
        setThingTypes(thingTypesData);
        setError('');
      } catch (err) {
        setError('Failed to fetch data. ' + err.message);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [user]);

  const handleEdit = (id) => {
    navigate(`/owned-things/edit/${id}`);
  };

  const handleDelete = async (id) => {
    if (window.confirm('Are you sure you want to delete this item?')) {
      try {
        await deleteOwnedThing(id, user.token);
        setOwnedThings(ownedThings.filter(thing => thing.owned_thing_id !== id));
      } catch (err) {
        setError('Failed to delete. ' + err.message);
      }
    }
  };

  const handleCreate = () => {
    navigate('/owned-things/create');
  };

  if (loading) return <div>Loading...</div>;
  if (error) return <div className="error-message">{error}</div>;

  return (
    <div className="owned-things-container">
      <h1>My Owned Things</h1>
      
      {thingTypes.length === 0 ? (
        <div className="error-message">
          <p>There are no types of things to own, please add some.</p>
          <Link to="/thing-types" className="button-link">Go to Thing Types</Link>
        </div>
      ) : (
        <button onClick={handleCreate} className="create-button">Create New</button>
      )}
      
      {thingTypes.length > 0 && ownedThings.length === 0 ? (
        <p>You don't have any items yet. Click 'Create New' to add one.</p>
      ) : null}
      
      {ownedThings.length > 0 && (
        <div className="owned-things-list">
          {ownedThings.map((thing) => (
            <div key={thing.owned_thing_id} className="owned-thing-card">
              <h3>{thing.name}</h3>
              <p>Type: {thing.thing_type ? thing.thing_type.name : 'Unknown'}</p>
              <p>Created: {new Date(thing.created_at).toLocaleDateString()}</p>
              <div className="card-actions">
                <button onClick={() => handleEdit(thing.owned_thing_id)}>Edit</button>
                <button onClick={() => handleDelete(thing.owned_thing_id)} className="delete-button">Delete</button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

export default OwnedThingsList;
