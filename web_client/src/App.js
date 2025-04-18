import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import Sidebar from './components/Sidebar';
import Home from './components/Home';
import Login from './components/Login';
import Signup from './components/Signup';
import OwnedThingsList from './components/OwnedThingsList';
import CreateOwnedThing from './components/CreateOwnedThing';
import EditOwnedThing from './components/EditOwnedThing';
import ThingTypesList from './components/ThingTypesList';
import CreateThingType from './components/CreateThingType';
import EditThingType from './components/EditThingType';
import { useAuth } from './contexts/AuthContext';

// Protected route component
const ProtectedRoute = ({ children }) => {
  const { isAuthenticated, loading } = useAuth();
  
  if (loading) {
    return <div>Loading...</div>;
  }
  
  if (!isAuthenticated) {
    return <Navigate to="/login" />;
  }
  
  return children;
};

function App() {
  return (
    <Router>
      <div className="app">
        <Sidebar />
        <div className="content">
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/login" element={<Login />} />
            <Route path="/signup" element={<Signup />} />
            <Route 
              path="/owned-things" 
              element={
                <ProtectedRoute>
                  <OwnedThingsList />
                </ProtectedRoute>
              } 
            />
            <Route 
              path="/owned-things/create" 
              element={
                <ProtectedRoute>
                  <CreateOwnedThing />
                </ProtectedRoute>
              } 
            />
            <Route 
              path="/owned-things/edit/:id" 
              element={
                <ProtectedRoute>
                  <EditOwnedThing />
                </ProtectedRoute>
              } 
            />
            <Route 
              path="/thing-types" 
              element={
                <ProtectedRoute>
                  <ThingTypesList />
                </ProtectedRoute>
              } 
            />
            <Route 
              path="/thing-types/create" 
              element={
                <ProtectedRoute>
                  <CreateThingType />
                </ProtectedRoute>
              } 
            />
            <Route 
              path="/thing-types/edit/:id" 
              element={
                <ProtectedRoute>
                  <EditThingType />
                </ProtectedRoute>
              } 
            />
          </Routes>
        </div>
      </div>
    </Router>
  );
}

export default App;
