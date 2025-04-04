import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { getOwnedThingById, updateOwnedThing } from '../services/ownedThingService';

function EditOwnedThing() {
  const [name, setName] = useState('');
  const [thingType, setThingType] = useState('');
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const { user } = useAuth();
  const navigate = useNavigate();
  const { id } = useParams();

  useEffect(() => {
    const fetchOwnedThing = async () => {
      if (!user || !id) return;
      
      try {
        const data = await getOwnedThingById(id, user.token);
        setName(data.name);
        setThingType(data.thing_type ? data.thing_type.name : 'Unknown');
        setError('');
      } catch (err) {
        setError('Failed to fetch owned thing. ' + err.message);
      } finally {
        setLoading(false);
      }
    };

    fetchOwnedThing();
  }, [user, id]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    
    if (!name.trim()) {
      setError('Name is required');
      return;
    }
    
    try {
      setSaving(true);
      await updateOwnedThing(id, { name }, user.token);
      navigate('/owned-things');
    } catch (err) {
      setError('Failed to update owned thing. ' + err.message);
      setSaving(false);
    }
  };

  if (loading) return <div>Loading...</div>;

  return (
    <div className="form-container">
      <h2>Edit Owned Thing</h2>
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
          <label>Thing Type</label>
          <p className="read-only-field">{thingType}</p>
          <small className="field-hint">Thing type cannot be changed</small>
        </div>
        
        <div className="form-actions">
          <button type="button" onClick={() => navigate('/owned-things')} className="cancel-button">
            Cancel
          </button>
          <button type="submit" disabled={saving}>
            {saving ? 'Saving...' : 'Save Changes'}
          </button>
        </div>
      </form>
    </div>
  );
}

export default EditOwnedThing;
