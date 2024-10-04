// src/App.tsx
import React from 'react';
import LogForm from './components/LogForm';

const App: React.FC = () => {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-100">
      <LogForm />
    </div>
  );
};

export default App;
