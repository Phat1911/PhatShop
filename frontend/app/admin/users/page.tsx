'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/lib/api';
import dayjs from 'dayjs';
import toast from 'react-hot-toast';
import clsx from 'clsx';

interface AdminUser {
  id: string;
  username: string;
  email: string;
  display_name: string;
  role: string;
  is_active: boolean;
  created_at: string;
}

export default function AdminUsersPage() {
  const qc = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['admin-users'],
    queryFn: () => api.get('/admin/users', { params: { limit: 50 } }).then((r) => r.data),
  });

  const updateRole = useMutation({
    mutationFn: ({ id, role }: { id: string; role: string }) =>
      api.patch(`/admin/users/${id}/role`, { role }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin-users'] }); toast.success('Đã cập nhật role'); },
    onError: () => toast.error('Cập nhật thất bại'),
  });

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Người dùng</h1>
      {isLoading ? (
        <div className="space-y-2">{[...Array(5)].map((_, i) => <div key={i} className="h-12 bg-gray-100 rounded-lg animate-pulse" />)}</div>
      ) : (
        <div className="bg-white rounded-xl border border-gray-100 overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-gray-50 border-b border-gray-100">
              <tr>
                <th className="text-left px-4 py-3 font-medium text-gray-600">Tên đăng nhập</th>
                <th className="text-left px-4 py-3 font-medium text-gray-600">Email</th>
                <th className="text-left px-4 py-3 font-medium text-gray-600">Ngày tạo</th>
                <th className="text-center px-4 py-3 font-medium text-gray-600">Role</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {(data?.data || []).map((u: AdminUser) => (
                <tr key={u.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 font-medium text-gray-900">{u.username}</td>
                  <td className="px-4 py-3 text-gray-600">{u.email}</td>
                  <td className="px-4 py-3 text-gray-500">{dayjs(u.created_at).format('DD/MM/YYYY')}</td>
                  <td className="px-4 py-3 text-center">
                    <select
                      value={u.role}
                      onChange={(e) => updateRole.mutate({ id: u.id, role: e.target.value })}
                      className={clsx(
                        'text-xs border rounded px-2 py-1 focus:outline-none',
                        u.role === 'admin' ? 'border-indigo-300 text-indigo-700' : 'border-gray-200 text-gray-600'
                      )}
                    >
                      <option value="user">user</option>
                      <option value="admin">admin</option>
                    </select>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
