import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import '@gke-mcp/ui/shared/styles/variables.css';
import '@gke-mcp/ui/shared/styles/core.css';
import './index.css';
import App from './App.tsx';

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>
);
