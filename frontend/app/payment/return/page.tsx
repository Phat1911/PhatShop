'use client';

import { useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { FiCheckCircle, FiXCircle } from 'react-icons/fi';
import { Suspense } from 'react';

function ReturnContent() {
  const params = useSearchParams();
  const responseCode = params.get('vnp_ResponseCode');
  const txnRef = params.get('vnp_TxnRef');
  const success = responseCode === '00';

  return (
    <div className="min-h-screen flex items-center justify-center px-4">
      <div className="w-full max-w-sm bg-white rounded-2xl shadow-sm border border-gray-100 p-8 text-center">
        {success ? (
          <>
            <FiCheckCircle size={56} className="text-emerald-500 mx-auto mb-4" />
            <h1 className="text-xl font-bold text-gray-900 mb-2">Thanh toán thành công!</h1>
            <p className="text-sm text-gray-500 mb-6">Đơn hàng của bạn đã được xác nhận. Bạn có thể tải xuống sản phẩm ngay bây giờ.</p>
            <Link href="/orders" className="block w-full bg-indigo-600 text-white py-2.5 rounded-lg text-sm font-medium hover:bg-indigo-700">
              Xem đơn hàng
            </Link>
          </>
        ) : (
          <>
            <FiXCircle size={56} className="text-red-500 mx-auto mb-4" />
            <h1 className="text-xl font-bold text-gray-900 mb-2">Thanh toán thất bại</h1>
            <p className="text-sm text-gray-500 mb-2">Mã lỗi: {responseCode}</p>
            <p className="text-sm text-gray-500 mb-6">Vui lòng thử lại hoặc liên hệ hỗ trợ.</p>
            <Link href="/cart" className="block w-full bg-indigo-600 text-white py-2.5 rounded-lg text-sm font-medium hover:bg-indigo-700">
              Quay lại giỏ hàng
            </Link>
          </>
        )}
        {txnRef && <p className="text-xs text-gray-400 mt-4">Mã giao dịch: {txnRef}</p>}
      </div>
    </div>
  );
}

export default function PaymentReturnPage() {
  return <Suspense><ReturnContent /></Suspense>;
}
