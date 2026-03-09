'use client';

import { useAuthStore } from '@/lib/store';
import { useRouter } from 'next/navigation';
import { useEffect } from 'react';

export default function AdminGuard({ children }: { children: React.ReactNode }) {
  const user = useAuthStore((s) => s.user);
  const router = useRouter();

  useEffect(() => {
    if (user === null) {
      router.push('/auth/login');
    } else if (user.role !== 'admin') {
      router.push('/');
    }
  }, [user, router]);

  if (!user || user.role !== 'admin') return null;
  return <>{children}</>;
}
