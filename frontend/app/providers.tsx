'use client';

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Toaster } from 'react-hot-toast';
import { useEffect } from 'react';
import Cookies from 'js-cookie';
import { useAuthStore } from '@/lib/store';
import { api } from '@/lib/api';

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: 1, staleTime: 30000 } },
});

export default function Providers({ children }: { children: React.ReactNode }) {
  const clearAuth = useAuthStore((s) => s.clearAuth);
  const setAuth = useAuthStore((s) => s.setAuth);
  const setLoading = useAuthStore((s) => s.setLoading);

  useEffect(() => {
    const token = Cookies.get('phatshop_token');
    if (token) {
      api.get('/users/me').then((res) => {
        setAuth(res.data, token);
      }).catch(() => {
        clearAuth();
      });
    } else {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    const handle = () => {
      clearAuth();
      window.location.href = '/auth/login';
    };
    window.addEventListener('phatshop:auth-expired', handle);
    return () => window.removeEventListener('phatshop:auth-expired', handle);
  }, [clearAuth]);

  return (
    <QueryClientProvider client={queryClient}>
      {children}
      <Toaster position="top-right" />
    </QueryClientProvider>
  );
}
