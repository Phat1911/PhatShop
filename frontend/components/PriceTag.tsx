interface PriceTagProps {
  amount: number;
  className?: string;
}

export default function PriceTag({ amount, className = '' }: PriceTagProps) {
  const formatted = new Intl.NumberFormat('vi-VN', {
    style: 'currency',
    currency: 'VND',
  }).format(amount);
  return <span className={className}>{formatted}</span>;
}
