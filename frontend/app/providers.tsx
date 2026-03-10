'use client';

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Toaster } from 'react-hot-toast';
import { useEffect } from 'react';
import Cookies from 'js-cookie';
import { useAuthStore } from '@/lib/store';
import { api } from '@/lib/api';

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: 1, staleTime: 0 } },
});

function AuthInitializer() {
  const { setAuth, clearAuth, setLoading } = useAuthStore();

  useEffect(() => {
    const initAuth = async () => {
      const token = Cookies.get('phatshop_token');
      if (!token) {
        setLoading(false);
        return;
      }
      try {
        const { data } = await api.get('/users/me');
        setAuth(data, token);
      } catch {
        clearAuth();
      }
    };

    initAuth();
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    const handle = () => {
      clearAuth();
      window.location.href = '/auth/login';
    };
    window.addEventListener('phatshop:auth-expired', handle);
    return () => window.removeEventListener('phatshop:auth-expired', handle);
  }, [clearAuth]);

  return null;
}

export default function Providers({ children }: { children: React.ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthInitializer />
      {children}
      <Toaster position="top-right" />
    </QueryClientProvider>
  );
}
