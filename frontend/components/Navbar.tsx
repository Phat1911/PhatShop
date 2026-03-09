'use client';

import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { FiShoppingCart, FiUser, FiLogOut, FiPackage, FiSettings } from 'react-icons/fi';
import { useAuthStore, useCartStore } from '@/lib/store';
import { useEffect, useRef, useState } from 'react';
import { api } from '@/lib/api';

export default function Navbar() {
  const { user, clearAuth } = useAuthStore();
  const { items, setItems } = useCartStore();
  const router = useRouter();
  const [menuOpen, setMenuOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (user) {
      api.get('/cart').then((r) => setItems(r.data.items || [])).catch(() => {});
    }
  }, [user, setItems]);

  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, []);

  const logout = () => {
    clearAuth();
    setMenuOpen(false);
    router.push('/');
  };

  return (
    <nav className="bg-white border-b border-gray-200 sticky top-0 z-50">
      <div className="max-w-7xl mx-auto px-4 h-16 flex items-center justify-between">
        <Link href="/products" className="text-xl font-bold text-indigo-600 tracking-tight">
          PhatShop
        </Link>

        <div className="flex items-center gap-4">
          <Link href="/products" className="text-sm text-gray-600 hover:text-gray-900">
            Sản phẩm
          </Link>

          <Link href="/cart" className="relative text-gray-600 hover:text-gray-900">
            <FiShoppingCart size={22} />
            {items.length > 0 && (
              <span className="absolute -top-1.5 -right-1.5 bg-indigo-600 text-white text-xs rounded-full w-4 h-4 flex items-center justify-center">
                {items.length}
              </span>
            )}
          </Link>

          {user ? (
            <div className="relative" ref={menuRef}>
              <button
                onClick={() => setMenuOpen((o) => !o)}
                className="flex items-center gap-2 text-sm text-gray-700 hover:text-gray-900"
              >
                <FiUser size={20} />
                <span className="hidden sm:inline max-w-[100px] truncate">{user.display_name || user.username}</span>
              </button>
              {menuOpen && (
                <div className="absolute right-0 mt-2 w-48 bg-white rounded-lg shadow-lg border border-gray-100 py-1 z-50">
                  <Link href="/orders" onClick={() => setMenuOpen(false)}
                    className="flex items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50">
                    <FiPackage size={16} /> Đơn hàng
                  </Link>
                  {user.role === 'admin' && (
                    <Link href="/admin" onClick={() => setMenuOpen(false)}
                      className="flex items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50">
                      <FiSettings size={16} /> Quản trị
                    </Link>
                  )}
                  <hr className="my-1 border-gray-100" />
                  <button onClick={logout}
                    className="flex items-center gap-2 w-full px-4 py-2 text-sm text-red-600 hover:bg-gray-50">
                    <FiLogOut size={16} /> Đăng xuất
                  </button>
                </div>
              )}
            </div>
          ) : (
            <Link href="/auth/login"
              className="bg-indigo-600 text-white text-sm px-4 py-1.5 rounded-lg hover:bg-indigo-700">
              Đăng nhập
            </Link>
          )}
        </div>
      </div>
    </nav>
  );
}
