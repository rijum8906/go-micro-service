import { useAuthStore } from '@/store/auth';
import axios from 'axios';

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 5000,
});

// Request Interceptor: Attach Token automatically
api.interceptors.request.use(
  (config) => {
    // Access the store state directly (outside of React lifecycle)
    const token = useAuthStore.getState().token;

    if (token) {
      config.headers.Authorization = `Bearer ${token.accessToken}`;
    }

    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);
