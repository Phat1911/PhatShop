'use client';

import Link from 'next/link';

import { FiShoppingCart, FiUser, FiLogOut, FiPackage, FiSettings } from 'react-icons/fi';
import { useAuthStore, useCartStore } from '@/lib/store';
import { useEffect, useRef, useState } from 'react';
import { api } from '@/lib/api';

export default function Navbar() {
  const { user, clearAuth } = useAuthStore();
  const { items, setItems } = useCartStore();
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
    window.location.href = '/';
  };

  return (
    <nav className="glass border-b border-[#1f1f2e] sticky top-0 z-50">
      <div className="max-w-7xl mx-auto px-4 h-16 flex items-center justify-between">
        <Link href="/products" className="text-xl font-bold tracking-tight">
          <span className="text-[#e63946]">Phat</span><span className="text-white">Shop</span>
        </Link>

        <div className="flex items-center gap-4">
          <Link href="/products" className="text-sm text-gray-400 hover:text-white transition-colors">
            Sản phẩm
          </Link>

          <Link href="/cart" className="relative text-gray-400 hover:text-white transition-colors">
            <FiShoppingCart size={22} />
            {items.length > 0 && (
              <span className="absolute -top-1.5 -right-1.5 bg-[#e63946] text-white text-xs rounded-full w-4 h-4 flex items-center justify-center animate-pulse-glow">
                {items.length}
              </span>
            )}
          </Link>

          {user ? (
            <div className="relative" ref={menuRef}>
              <button
                onClick={() => setMenuOpen((o) => !o)}
                className="flex items-center gap-2 text-sm text-gray-300 hover:text-white transition-colors"
              >
                <FiUser size={20} />
                <span className="hidden sm:inline max-w-[100px] truncate">{user.display_name || user.username}</span>
              </button>
              {menuOpen && (
                <div className="absolute right-0 mt-2 w-48 bg-[#16161e] rounded-lg shadow-2xl border border-[#1f1f2e] py-1 z-50 animate-slide-down">
                  <Link href="/orders" onClick={() => setMenuOpen(false)}
                    className="flex items-center gap-2 px-4 py-2 text-sm text-gray-300 hover:bg-[#1f1f2e] hover:text-white transition-colors">
                    <FiPackage size={16} /> Đơn hàng
                  </Link>
                  {user.role === 'admin' && (
                    <Link href="/admin" onClick={() => setMenuOpen(false)}
                      className="flex items-center gap-2 px-4 py-2 text-sm text-gray-300 hover:bg-[#1f1f2e] hover:text-white transition-colors">
                      <FiSettings size={16} /> Quản trị
                    </Link>
                  )}
                  <hr className="my-1 border-[#1f1f2e]" />
                  <button onClick={logout}
                    className="flex items-center gap-2 w-full px-4 py-2 text-sm text-[#e63946] hover:bg-[#1f1f2e] transition-colors">
                    <FiLogOut size={16} /> Đăng xuất
                  </button>
                </div>
              )}
            </div>
          ) : (
            <Link href="/auth/login"
              className="bg-[#e63946] hover:bg-[#ff4d5a] text-white text-sm px-4 py-1.5 rounded-lg transition-colors font-medium">
              Đăng nhập
            </Link>
          )}
        </div>
      </div>
    </nav>
  );
}
