import AdminGuard from '@/components/AdminGuard';
import Link from 'next/link';

const NAV = [
  { href: '/admin', label: 'Thống kê' },
  { href: '/admin/products', label: 'Sản phẩm' },
  { href: '/admin/orders', label: 'Đơn hàng' },
  { href: '/admin/users', label: 'Người dùng' },
];

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  return (
    <AdminGuard>
      <div className="flex min-h-screen">
        <aside className="w-52 bg-white border-r border-gray-200 flex-shrink-0">
          <div className="p-4 border-b border-gray-100">
            <p className="text-xs text-gray-500 uppercase font-semibold tracking-wider">Quản trị</p>
          </div>
          <nav className="p-3 space-y-1">
            {NAV.map((n) => (
              <Link key={n.href} href={n.href}
                className="block px-3 py-2 text-sm text-gray-700 hover:bg-indigo-50 hover:text-indigo-700 rounded-lg">
                {n.label}
              </Link>
            ))}
          </nav>
        </aside>
        <div className="flex-1 p-6 overflow-auto">{children}</div>
      </div>
    </AdminGuard>
  );
}
