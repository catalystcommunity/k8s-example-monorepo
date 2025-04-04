import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { getThingTypeById, updateThingType } from '../services/thingTypeService';

function EditThingType() {
  const { id } = useParams();
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const { user } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    const fetchThingType = async () => {
      if (!user || !id) return;
      
      try {
        setLoading(true);
        const data = await getThingTypeById(id, user.token);
        setName(data.name);
        setDescription(data.description || '');
        setError('');
      } catch (err) {
        setError('Failed to fetch thing type. ' + err.message);
      } finally {
        setLoading(false);
      }
    };

    fetchThingType();
  }, [id, user]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    
    if (!name.trim()) {
      setError('Name is required');
      return;
    }
    
    try {
      setSaving(true);
      await updateThingType(id, {
        name,
        description
      }, user.token);
      
      navigate('/thing-types');
    } catch (err) {
      setError('Failed to update thing type. ' + err.message);
      setSaving(false);
    }
  };

  if (loading) return <div>Loading...</div>;

  return (
    <div className="form-container">
      <h2>Edit Thing Type</h2>
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
          <label htmlFor="description">Description</label>
          <textarea
            id="description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows="4"
          />
        </div>
        
        <div className="form-actions">
          <button type="button" onClick={() => navigate('/thing-types')} className="cancel-button">
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

export default EditThingType;