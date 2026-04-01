import React from 'react';
import ReactDOM from 'react-dom/client';
import '@gke-mcp/ui/shared/styles/variables.css';
import '@gke-mcp/ui/shared/styles/core.css';
import App from './App';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
