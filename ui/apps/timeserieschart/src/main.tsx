import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import './index.css';
import '@gke-mcp/ui/shared/styles/variables.css';
import App from './App.tsx';

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>
);
