import axios from 'axios';
import Cookies from 'js-cookie';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

export const UPLOADS_URL = process.env.NEXT_PUBLIC_UPLOADS_URL || 'http://localhost:8080';

export const api = axios.create({ baseURL: API_URL });

api.interceptors.request.use((config) => {
  const token = Cookies.get('phatshop_token');
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      Cookies.remove('phatshop_token');
      if (typeof window !== 'undefined') {
        window.dispatchEvent(new CustomEvent('phatshop:auth-expired'));
      }
    }
    return Promise.reject(err);
  }
);

export const getUploadUrl = (path: string) => {
  if (!path) return '';
  if (path.startsWith('http')) return path;
  return UPLOADS_URL + path;
};
