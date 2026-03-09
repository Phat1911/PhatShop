'use client';

import { useQuery, useMutation } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { useParams, useRouter } from 'next/navigation';
import { useCartStore } from '@/lib/store';
import Image from 'next/image';
import PriceTag from '@/components/PriceTag';
import { FiCopy, FiCheckCircle, FiUpload, FiXCircle, FiAlertCircle } from 'react-icons/fi';
import { useEffect, useState, useRef } from 'react';
import toast from 'react-hot-toast';
import Link from 'next/link';

const BANK_ID = 'VPB';
const ACCOUNT_NO = '0764717493';
const ACCOUNT_NAME = 'TRAN DINH HONG PHAT';

function buildQRUrl(amount: number, note: string) {
  const encodedNote = encodeURIComponent(note);
  const encodedName = encodeURIComponent(ACCOUNT_NAME);
  return `https://img.vietqr.io/image/${BANK_ID}-${ACCOUNT_NO}-compact2.png?amount=${amount}&addInfo=${encodedNote}&accountName=${encodedName}`;
}

export default function PaymentQRPage() {
  const { orderId } = useParams<{ orderId: string }>();
  const router = useRouter();
  const clearCart = useCartStore((s) => s.clearCart);
  const [copied, setCopied] = useState<string | null>(null);
  const [receiptFile, setReceiptFile] = useState<File | null>(null);
  const [receiptPreview, setReceiptPreview] = useState<string | null>(null);
  const [verifyResult, setVerifyResult] = useState<{ success: boolean; message: string } | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const uploadReceipt = useMutation({
    mutationFn: async (file: File) => {
      const form = new FormData();
      form.append('receipt', file);
      const { data } = await api.post(`/orders/${orderId}/receipt`, form, {
        headers: { 'Content-Type': 'multipart/form-data' },
      });
      return data;
    },
    onSuccess: (data) => {
      setVerifyResult({ success: true, message: data.message });
      toast.success('Thanh toán đã được xác nhận!');
      clearCart();
    },
    onError: (err: unknown) => {
      const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error
        || 'Xác nhận thất bại. Vui lòng thử lại.';
      setVerifyResult({ success: false, message: msg });
    },
  });

  const transferNote = `PHATSHOP ${orderId.slice(0, 8).toUpperCase()}`;

  const { data: order } = useQuery({
    queryKey: ['order-payment', orderId],
    queryFn: () => api.get(`/orders/${orderId}`).then((r) => r.data),
    refetchInterval: (query) => {
      if (query.state.data?.status === 'paid') return false;
      return 5000;
    },
  });

  useEffect(() => {
    if (order?.status === 'paid') {
      clearCart();
      toast.success('Thanh toán thành công! Bạn có thể tải xuống sản phẩm.');
      router.push(`/orders/${orderId}`);
    }
  }, [order?.status, orderId, router, clearCart]);

  const copy = (text: string, key: string) => {
    navigator.clipboard.writeText(text);
    setCopied(key);
    setTimeout(() => setCopied(null), 2000);
  };

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    setReceiptFile(file);
    setVerifyResult(null);
    const url = URL.createObjectURL(file);
    setReceiptPreview(url);
  };

  if (!order) {
    return (
      <div className="max-w-md mx-auto px-4 py-12">
        <div className="h-96 bg-gray-100 rounded-2xl animate-pulse" />
      </div>
    );
  }

  const qrUrl = buildQRUrl(order.total_amount, transferNote);

  return (
    <div className="max-w-md mx-auto px-4 py-8">
      <div className="bg-white rounded-2xl shadow-sm border border-gray-100 overflow-hidden">
        {/* Header */}
        <div className="bg-indigo-600 px-6 py-4 text-white text-center">
          <h1 className="text-lg font-bold">Thanh toán chuyển khoản</h1>
          <p className="text-indigo-200 text-sm mt-0.5">
            Mã đơn: #{orderId.slice(0, 8).toUpperCase()}
          </p>
        </div>

        {/* QR Code */}
        <div className="flex flex-col items-center px-6 pt-6 pb-4">
          <div className="border-2 border-indigo-100 rounded-xl p-2 bg-white shadow-sm">
            <Image
              src={qrUrl}
              alt="VietQR VPBank"
              width={240}
              height={290}
              className="rounded-lg"
              unoptimized
            />
          </div>
          <p className="text-xs text-gray-400 mt-2">Quét bằng app ngân hàng bất kỳ</p>
        </div>

        {/* Transfer details */}
        <div className="px-6 pb-6 space-y-3">
          <div className="bg-gray-50 rounded-xl divide-y divide-gray-100">
            <div className="flex justify-between items-center px-4 py-3">
              <span className="text-sm text-gray-500">Ngân hàng</span>
              <span className="text-sm font-semibold text-gray-900">VPBank</span>
            </div>
            <div className="flex justify-between items-center px-4 py-3">
              <span className="text-sm text-gray-500">Số tài khoản</span>
              <div className="flex items-center gap-2">
                <span className="text-sm font-semibold text-gray-900 font-mono">{ACCOUNT_NO}</span>
                <button onClick={() => copy(ACCOUNT_NO, 'acc')} className="text-indigo-500 hover:text-indigo-700">
                  {copied === 'acc' ? <FiCheckCircle size={15} /> : <FiCopy size={15} />}
                </button>
              </div>
            </div>
            <div className="flex justify-between items-center px-4 py-3">
              <span className="text-sm text-gray-500">Chủ tài khoản</span>
              <span className="text-sm font-semibold text-gray-900">{ACCOUNT_NAME}</span>
            </div>
            <div className="flex justify-between items-center px-4 py-3">
              <span className="text-sm text-gray-500">Số tiền</span>
              <div className="flex items-center gap-2">
                <PriceTag amount={order.total_amount} className="text-sm font-bold text-indigo-600" />
                <button onClick={() => copy(String(order.total_amount), 'amount')} className="text-indigo-500 hover:text-indigo-700">
                  {copied === 'amount' ? <FiCheckCircle size={15} /> : <FiCopy size={15} />}
                </button>
              </div>
            </div>
            <div className="flex justify-between items-center px-4 py-3">
              <span className="text-sm text-gray-500">Nội dung CK</span>
              <div className="flex items-center gap-2">
                <span className="text-sm font-bold text-red-600 font-mono">{transferNote}</span>
                <button onClick={() => copy(transferNote, 'note')} className="text-indigo-500 hover:text-indigo-700">
                  {copied === 'note' ? <FiCheckCircle size={15} /> : <FiCopy size={15} />}
                </button>
              </div>
            </div>
          </div>

          <p className="text-xs text-red-500 text-center font-medium">
            Ghi đúng nội dung chuyển khoản để đơn hàng được xử lý nhanh
          </p>

          {/* Receipt upload */}
          <div className="border-t border-gray-100 pt-4">
            <p className="text-sm font-semibold text-gray-700 mb-2">
              Chụp màn hình biên lai giao dịch để hệ thống tự xác nhận tức thì.
            </p>

            {verifyResult?.success ? (
              <div className="bg-emerald-50 border border-emerald-200 rounded-xl p-4 text-center">
                <FiCheckCircle className="text-emerald-500 mx-auto mb-2" size={32} />
                <p className="text-sm font-semibold text-emerald-700">Xác nhận thành công!</p>
                <p className="text-xs text-emerald-600 mt-1 mb-3">{verifyResult.message}</p>
                <Link
                  href={`/orders/${orderId}`}
                  className="inline-block bg-emerald-600 hover:bg-emerald-700 text-white text-sm font-medium px-5 py-2 rounded-lg"
                >
                  Xem và tải xuống ngay
                </Link>
              </div>
            ) : (
              <>
                {verifyResult?.success === false && (
                  <div className="flex items-start gap-2 bg-red-50 border border-red-200 rounded-xl px-4 py-3 mb-3">
                    <FiAlertCircle className="text-red-500 flex-shrink-0 mt-0.5" size={16} />
                    <p className="text-xs text-red-700">{verifyResult.message}</p>
                  </div>
                )}

                {receiptPreview && (
                  <div className="relative mb-3">
                    <img src={receiptPreview} alt="Receipt preview" className="w-full max-h-48 object-contain rounded-xl border border-gray-200" />
                    <button
                      onClick={() => { setReceiptFile(null); setReceiptPreview(null); setVerifyResult(null); }}
                      className="absolute top-2 right-2 bg-white rounded-full p-1 shadow text-gray-500 hover:text-red-500"
                    >
                      <FiXCircle size={16} />
                    </button>
                  </div>
                )}

                <input
                  ref={fileInputRef}
                  type="file"
                  accept="image/jpeg,image/png,image/webp"
                  capture="environment"
                  className="hidden"
                  onChange={handleFileSelect}
                />
                <button
                  onClick={() => fileInputRef.current?.click()}
                  className="w-full flex items-center justify-center gap-2 border-2 border-dashed border-gray-300 hover:border-indigo-400 rounded-xl py-3 text-sm text-gray-500 hover:text-indigo-600 transition-colors"
                >
                  <FiUpload size={16} />
                  {receiptFile ? receiptFile.name : 'Chọn ảnh biên lai hoặc chụp màn hình'}
                </button>

                {receiptFile && (
                  <button
                    onClick={() => uploadReceipt.mutate(receiptFile)}
                    disabled={uploadReceipt.isPending}
                    className="mt-3 w-full bg-indigo-600 hover:bg-indigo-700 disabled:opacity-60 text-white text-sm font-semibold py-2.5 rounded-xl transition-colors"
                  >
                    {uploadReceipt.isPending ? 'Đang phân tích biên lai...' : 'Xác nhận thanh toán'}
                  </button>
                )}
              </>
            )}
          </div>

          {/* Order items summary */}
          <div className="mt-2">
            <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Sản phẩm đặt mua</p>
            <div className="space-y-1">
              {order.items?.map((item: { product_id: string; price: number; product?: { title: string } }) => (
                <div key={item.product_id} className="flex justify-between text-sm">
                  <span className="text-gray-700 line-clamp-1 max-w-[220px]">{item.product?.title}</span>
                  <PriceTag amount={item.price} className="text-gray-600 whitespace-nowrap ml-2" />
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>

      <p className="text-center text-xs text-gray-400 mt-4">
        Sau khi chuyển khoản, trang sẽ tự xác nhận trong vài phút.
        <br />Hoặc liên hệ admin nếu cần hỗ trợ.
      </p>
    </div>
  );
}
