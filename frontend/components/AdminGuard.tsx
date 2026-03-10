'use client';

import { useAuthStore } from '@/lib/store';
import { useRouter } from 'next/navigation';
import { useEffect } from 'react';

export default function AdminGuard({ children }: { children: React.ReactNode }) {
  const user = useAuthStore((s) => s.user);
  const isLoading = useAuthStore((s) => s.isLoading);
  const router = useRouter();

  useEffect(() => {
    if (isLoading) return;
    if (user === null) {
      router.push('/auth/login');
    } else if (user.role !== 'admin') {
      router.push('/');
    }
  }, [user, isLoading, router]);

  if (isLoading) return null;
  if (!user || user.role !== 'admin') return null;
  return <>{children}</>;
}
