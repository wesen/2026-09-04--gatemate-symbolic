import React from 'react';
import { createRoot } from 'react-dom/client';
import { Provider } from 'react-redux';
import 'bootstrap/dist/css/bootstrap.min.css';
import './style.css';
import App from './App';
import { makeStore } from './store';

createRoot(document.getElementById('root')!).render(<React.StrictMode><Provider store={makeStore()}><App /></Provider></React.StrictMode>);
