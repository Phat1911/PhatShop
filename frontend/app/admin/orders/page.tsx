'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/lib/api';
import PriceTag from '@/components/PriceTag';
import dayjs from 'dayjs';
import clsx from 'clsx';
import toast from 'react-hot-toast';

const STATUS_MAP: Record<string, { label: string; cls: string }> = {
  pending: { label: 'Chờ TT', cls: 'bg-yellow-100 text-yellow-700' },
  paid:    { label: 'Đã TT',  cls: 'bg-emerald-100 text-emerald-700' },
  cancelled: { label: 'Đã huỷ', cls: 'bg-gray-100 text-gray-600' },
};

export default function AdminOrdersPage() {
  const qc = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['admin-orders'],
    queryFn: () => api.get('/admin/orders', { params: { limit: 50 } }).then((r) => r.data),
  });

  const updateStatus = useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) =>
      api.patch(`/admin/orders/${id}/status`, { status }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin-orders'] }); toast.success('Đã cập nhật'); },
    onError: () => toast.error('Cập nhật thất bại'),
  });

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Đơn hàng</h1>
      {isLoading ? (
        <div className="space-y-2">{[...Array(5)].map((_, i) => <div key={i} className="h-12 bg-gray-100 rounded-lg animate-pulse" />)}</div>
      ) : (
        <div className="bg-white rounded-xl border border-gray-100 overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-gray-50 border-b border-gray-100">
              <tr>
                <th className="text-left px-4 py-3 font-medium text-gray-600">Mã đơn</th>
                <th className="text-left px-4 py-3 font-medium text-gray-600">Ngày tạo</th>
                <th className="text-right px-4 py-3 font-medium text-gray-600">Tổng tiền</th>
                <th className="text-center px-4 py-3 font-medium text-gray-600">Trạng thái</th>
                <th className="text-center px-4 py-3 font-medium text-gray-600">Hành động</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {(data?.data || []).map((o: { id: string; created_at: string; total_amount: number; status: string }) => {
                const s = STATUS_MAP[o.status] || STATUS_MAP.pending;
                return (
                  <tr key={o.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 font-mono text-gray-900">#{o.id.slice(0, 8).toUpperCase()}</td>
                    <td className="px-4 py-3 text-gray-600">{dayjs(o.created_at).format('DD/MM/YY HH:mm')}</td>
                    <td className="px-4 py-3 text-right"><PriceTag amount={o.total_amount} className="text-indigo-600 font-medium" /></td>
                    <td className="px-4 py-3 text-center">
                      <span className={clsx('text-xs px-2 py-0.5 rounded-full font-medium', s.cls)}>{s.label}</span>
                    </td>
                    <td className="px-4 py-3 text-center">
                      <select
                        value={o.status}
                        onChange={(e) => updateStatus.mutate({ id: o.id, status: e.target.value })}
                        className="text-xs border border-gray-200 rounded px-2 py-1 focus:outline-none"
                      >
                        <option value="pending">Chờ TT</option>
                        <option value="paid">Đã TT</option>
                        <option value="cancelled">Huỷ</option>
                      </select>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
